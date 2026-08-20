package confws_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xoctopus/x/codex"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/hack"
	"github.com/xoctopus/confx/pkg/confws"
	"github.com/xoctopus/confx/pkg/types"
)

func TestEndpoint_RunRequiresHooks(t *testing.T) {
	ep := &confws.Endpoint{}
	ep.SetDefault()
	ep.ListenAddr = "127.0.0.1:0"
	Expect(t, ep.Init(context.Background()), Succeed())
	err := ep.Run(context.Background())
	Expect(t, err, Failed())
	Expect(t, codex.IsCode(err, confws.ERROR__MISSING_REQUIRED_HOOKS), BeTrue())
}

func TestEndpoint_InitTwice(t *testing.T) {
	ep := &confws.Endpoint{}
	ep.SetDefault()
	Expect(t, ep.Init(context.Background()), Succeed())
	err := ep.Init(context.Background())
	Expect(t, err, Failed())
	Expect(t, codex.IsCode(err, confws.ERROR__ENDPOINT_INITIALIZED), BeTrue())
}

func TestEndpoint_RunTwice(t *testing.T) {
	ep := hack.NewWSEndpoint(t, func(ep *confws.Endpoint) {
		ep.SetMessageHandler(func(context.Context, confws.Client, int, []byte) {})
	})
	err := ep.Run(context.Background())
	Expect(t, err, Failed())
	Expect(t, codex.IsCode(err, confws.ERROR__ENDPOINT_INITIALIZED), BeTrue())
}

func TestEndpoint_MessagePush(t *testing.T) {
	got := make(chan string, 1)
	ep := hack.NewWSEndpoint(t, func(ep *confws.Endpoint) {
		ep.SetMessageHandler(func(ctx context.Context, cli confws.Client, typ int, data []byte) {
			Expect(t, typ, Equal(websocket.TextMessage))
			got <- string(data)
			Expect(t, cli.WriteText(ctx, []byte("pong:"+string(data))), Succeed())
		})
	})

	c := hack.DialWS(t, ep)
	Expect(t, c.WriteMessage(websocket.TextMessage, []byte("ping")), Succeed())

	select {
	case msg := <-got:
		Expect(t, msg, Equal("ping"))
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting message handler")
	}

	_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := c.ReadMessage()
	Expect(t, err, Succeed())
	Expect(t, string(data), Equal("pong:ping"))
}

func TestEndpoint_EstablishPullAndReadGuard(t *testing.T) {
	ready := make(chan confws.Client, 1)
	ep := hack.NewWSEndpoint(t, func(ep *confws.Endpoint) {
		ep.SetEstablishHandler(func(ctx context.Context, cli confws.Client) (context.Context, error) {
			ready <- cli
			return ctx, nil
		})
		ep.SetMessageHandler(func(context.Context, confws.Client, int, []byte) {})
	})

	c := hack.DialWS(t, ep)
	var cli confws.Client
	select {
	case cli = <-ready:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting establish")
	}

	_, _, err := cli.Read(context.Background())
	Expect(t, err, Failed())
	Expect(t, codex.IsCode(err, confws.ERROR__READ_ON_NON_PULL_CLIENT), BeTrue())

	_ = c
}

func TestEndpoint_PullRead(t *testing.T) {
	type result struct {
		typ  int
		data []byte
		err  error
	}
	out := make(chan result, 1)
	ep := hack.NewWSEndpoint(t, func(ep *confws.Endpoint) {
		ep.SetEstablishHandler(func(ctx context.Context, cli confws.Client) (context.Context, error) {
			go func() {
				typ, data, err := cli.Read(ctx)
				out <- result{typ, data, err}
			}()
			return ctx, nil
		})
	})

	c := hack.DialWS(t, ep)
	Expect(t, c.WriteMessage(websocket.BinaryMessage, []byte{1, 2, 3}), Succeed())

	select {
	case r := <-out:
		Expect(t, r.err, Succeed())
		Expect(t, r.typ, Equal(websocket.BinaryMessage))
		Expect(t, r.data, Equal([]byte{1, 2, 3}))
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting pull read")
	}
	_ = c
}

func TestEndpoint_ConnectionHandlerReject(t *testing.T) {
	ep := hack.NewWSEndpoint(t, func(ep *confws.Endpoint) {
		ep.SetConnectionHandler(func(ctx context.Context, r *http.Request) (context.Context, []confws.ClientOptionApplier, error) {
			Expect(t, r.URL.Path, Equal("/ws"))
			return ctx, nil, errors.New("unauthorized")
		})
		ep.SetEstablishHandler(func(ctx context.Context, _ confws.Client) (context.Context, error) { return ctx, nil })
	})

	_, resp, err := websocket.DefaultDialer.Dial(ep.String(), nil)
	Expect(t, err, Failed())
	if resp != nil {
		Expect(t, resp.StatusCode, Equal(http.StatusUnauthorized))
		_ = resp.Body.Close()
	}
}

func TestEndpoint_ConnectionHandlerAppliers(t *testing.T) {
	var disc atomic.Value
	established := make(chan string, 1)

	ep := hack.NewWSEndpoint(t, func(ep *confws.Endpoint) {
		ep.SetConnectionHandler(func(ctx context.Context, _ *http.Request) (context.Context, []confws.ClientOptionApplier, error) {
			return ctx, []confws.ClientOptionApplier{
				confws.WithClientDisconnectionHandler(func(_ context.Context, cid string, _ error) {
					disc.Store(cid)
				}),
				confws.WithClientEstablishHandler(func(ctx context.Context, cli confws.Client) (context.Context, error) {
					established <- cli.ID()
					return ctx, nil
				}),
			}, nil
		})
		ep.SetEstablishHandler(func(ctx context.Context, _ confws.Client) (context.Context, error) { return ctx, nil })
	})

	c := hack.DialWS(t, ep)
	var id string
	select {
	case id = <-established:
		Expect(t, id != "", BeTrue())
	case <-time.After(2 * time.Second):
		t.Fatal("timeout establish applier")
	}

	cli := ep.Client(id)
	Expect(t, cli, NotBeNil[confws.Client]())
	Expect(t, cli.Close(context.Background()), Succeed())

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if v := disc.Load(); v != nil {
			Expect(t, v.(string), Equal(id))
			_ = c
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("timeout waiting disconnect handler")
}

func TestEndpoint_MaxConnection(t *testing.T) {
	hold := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	ep := hack.NewWSEndpoint(t, func(ep *confws.Endpoint) {
		ep.MaxConnection = 1
		ep.SetEstablishHandler(func(ctx context.Context, _ confws.Client) (context.Context, error) {
			wg.Done()
			<-hold
			return ctx, nil
		})
	})

	c1 := hack.DialWS(t, ep)
	wg.Wait()

	_, resp, err := websocket.DefaultDialer.Dial(ep.String(), nil)
	Expect(t, err, Failed())
	if resp != nil {
		Expect(t, resp.StatusCode, Equal(http.StatusServiceUnavailable))
		_ = resp.Body.Close()
	}

	close(hold)
	_ = c1
}

func TestEndpoint_IdleTimeout(t *testing.T) {
	disc := make(chan error, 1)
	ep := hack.NewWSEndpoint(t, func(ep *confws.Endpoint) {
		ep.IdleTimeout = types.Duration(1500 * time.Millisecond)
		ep.SetConnectionHandler(func(ctx context.Context, _ *http.Request) (context.Context, []confws.ClientOptionApplier, error) {
			return ctx, []confws.ClientOptionApplier{
				confws.WithClientDisconnectionHandler(func(_ context.Context, _ string, err error) {
					disc <- err
				}),
			}, nil
		})
		ep.SetEstablishHandler(func(ctx context.Context, _ confws.Client) (context.Context, error) { return ctx, nil })
	})

	c := hack.DialWS(t, ep)
	select {
	case err := <-disc:
		Expect(t, codex.IsCode(err, confws.ERROR__CLIENT_IDLE_TIMEOUT), BeTrue())
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting idle close")
	}
	_ = c
}

func TestEndpoint_CloseRemovesClients(t *testing.T) {
	ready := make(chan confws.Client, 1)
	ep := hack.NewWSEndpoint(t, func(ep *confws.Endpoint) {
		ep.SetEstablishHandler(func(ctx context.Context, cli confws.Client) (context.Context, error) {
			ready <- cli
			return ctx, nil
		})
	})

	c := hack.DialWS(t, ep)
	var cli confws.Client
	select {
	case cli = <-ready:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout establish")
	}
	Expect(t, ep.Client(cli.ID()), NotBeNil[confws.Client]())

	Expect(t, ep.Close(context.Background()), Succeed())

	select {
	case <-cli.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("client not closed after endpoint Close")
	}
	Expect(t, ep.Client(cli.ID()), BeNil[confws.Client]())
	_ = c
}

func TestEndpoint_ClientWriteAfterClose(t *testing.T) {
	ready := make(chan confws.Client, 1)
	ep := hack.NewWSEndpoint(t, func(ep *confws.Endpoint) {
		ep.SetEstablishHandler(func(ctx context.Context, cli confws.Client) (context.Context, error) {
			ready <- cli
			return ctx, nil
		})
	})
	_ = hack.DialWS(t, ep)

	var cli confws.Client
	select {
	case cli = <-ready:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}

	Expect(t, cli.Close(context.Background()), Succeed())
	err := cli.WriteText(context.Background(), []byte("x"))
	Expect(t, err, Failed())
	Expect(t, codex.IsCode(err, confws.ERROR__CLIENT_CLOSED), BeTrue())
}

func TestEndpoint_EstablishErrorCloses(t *testing.T) {
	ep := hack.NewWSEndpoint(t, func(ep *confws.Endpoint) {
		ep.SetEstablishHandler(func(ctx context.Context, _ confws.Client) (context.Context, error) {
			return ctx, fmt.Errorf("reject establish")
		})
	})

	c, _, err := websocket.DefaultDialer.Dial(ep.String(), nil)
	if err != nil {
		return
	}
	defer c.Close()
	_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, err = c.ReadMessage()
	Expect(t, err, Failed())
}
