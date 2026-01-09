package confpulsar_test

import (
	"context"
	_ "embed"
	"strings"
	"testing"
	"time"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/hack"
	"github.com/xoctopus/confx/pkg/components/confmq"
	"github.com/xoctopus/confx/pkg/components/confpulsar/v1"
	"github.com/xoctopus/confx/pkg/components/conftls"
	"github.com/xoctopus/confx/pkg/types"
)

var (
	//go:embed testdata/client.key
	key string
	//go:embed testdata/client.crt
	crt string
	//go:embed testdata/ca.crt
	ca string
)

func genTopic(t *testing.T) string {
	return strings.ReplaceAll(t.Name(), "/", "_")
}

func TestPulsarEndpointV1(t *testing.T) {
	t.Run("SetDefault", func(t *testing.T) {
		ep := &confpulsar.Endpoint{}
		defer func() { _ = ep.Close() }()

		ep.SetDefault()
		_ = ep.Init(hack.Context(t))
		Expect(t, ep.Endpoint.Endpoint(), Equal("pulsar://localhost:6650"))
	})

	t.Run("Init", func(t *testing.T) {
		t.Run("Lost", func(t *testing.T) {
			t.Run("InvalidAddress", func(t *testing.T) {
				ep := &confpulsar.Endpoint{}
				ep.Address = "pulsar://localhost:6379/%zz"
				Expect(t, ep.Init(context.Background()), Failed())
			})

			t.Run("Unreachable", func(t *testing.T) {
				ctx := hack.WithPulsarLost(hack.Context(t), t, "pulsar://localhost:6650?connectionTimeout=100ms")
				ep := confmq.Must(ctx)
				Expect(t, ep, NotBeNil[confmq.PubSub]())

				_, err := ep.Subscribe(ctx, genTopic(t))
				Expect(t, err, Failed())

				err = ep.Publish(ctx, confmq.NewMessage(ctx, genTopic(t), "any"))
				Expect(t, err, Failed())
			})

			t.Run("WithTLS", func(t *testing.T) {
				ctx := hack.WithPulsarLost(hack.Context(t), t, "pulsar://localhost:6650")

				ep := &confpulsar.Endpoint{}
				ep.SetDefault()
				ep.Cert = conftls.X509KeyPair{Key: key, Crt: crt, CA: ca}

				Expect(t, ep.Init(ctx), Failed())
			})
		})

		dsn := "pulsar://localhost:16650"
		t.Run("Established", func(t *testing.T) {
			ctx := hack.WithPulsar(hack.Context(t), t, dsn)
			ep := confmq.Must(ctx)
			Expect(t, ep, NotBeNil[confmq.PubSub]())

			msg := confmq.NewMessage(ctx, genTopic(t), "any")
			ret := make(<-chan error)

			sub, err := ep.Subscribe(ctx, genTopic(t))
			Expect(t, err, Succeed())

			go func() {
				ret = sub.Run(ctx, func(ctx context.Context, rec confmq.Message) error {
					Expect(t, rec.Topic(), Equal(msg.Topic()))
					if rec.ID() == msg.ID() {
						time.Sleep(time.Second) // wait unacked messages
						_ = sub.Close()
					}
					return nil
				})
			}()

			time.Sleep(time.Second)
			err = ep.Publish(ctx, msg, confpulsar.WithSyncPublish())
			Expect(t, err, Succeed())
			t.Log(<-ret)

			t.Run("LivenessCheck", func(t *testing.T) {
				d := ep.(types.LivenessChecker).LivenessCheck(ctx)
				Expect(t, d.Reachable, BeTrue())
			})
		})

		t.Run("ClosedClient", func(t *testing.T) {
			ctx := hack.WithPulsar(hack.Context(t), t, dsn)
			ep := confmq.Must(ctx)
			_ = ep.Close()
			_, err := ep.Subscribe(ctx, genTopic(t))
			Expect(t, err, Failed())
			err = ep.Publish(ctx, confmq.NewMessage(ctx, genTopic(t), nil))
			Expect(t, err, Failed())
			r := ep.(types.LivenessChecker).LivenessCheck(ctx)
			Expect(t, r.Reachable, BeFalse())
			Expect(t, r.Message, Equal("endpoint is closed"))
		})

		t.Run("SendMode", func(t *testing.T) {
			t.Run("Sync", func(t *testing.T) {
				var (
					ctx = hack.WithPulsar(hack.Context(t), t, dsn)
					ep  = confmq.Must(ctx)
				)
				sub, err := ep.Subscribe(ctx, genTopic(t))
				Expect(t, err, Succeed())

				rec := sub.Run(
					ctx, func(ctx context.Context, msg confmq.Message) error {
						raw := string(msg.Data())
						Expect(t, raw, Equal("send_mode:sync"))
						time.Sleep(time.Second)
						_ = sub.Close()
						return nil
					},
				)
				err = ep.Publish(
					ctx, confmq.NewMessage(ctx, genTopic(t), "send_mode:sync"),
					confpulsar.WithSyncPublish(),
				)
				Expect(t, err, Succeed())
				t.Log(<-rec)
			})
			t.Run("Async", func(t *testing.T) {
				var (
					ctx = hack.WithPulsar(hack.Context(t), t, dsn)
					ep  = confmq.Must(ctx)
				)
				sub, err := ep.Subscribe(ctx, genTopic(t))
				Expect(t, err, Succeed())
				rec := sub.Run(ctx, func(ctx context.Context, msg confmq.Message) error {
					t.Log(string(msg.Data()))
					return nil
				})

				err = ep.Publish(
					ctx, confmq.NewMessage(ctx, genTopic(t), "send_mode:async"),
					confpulsar.WithPublishFailover(func(msg confmq.Message, err error) {
						raw := string(msg.Data())
						Expect(t, raw, Equal("send_mode:async"))
						time.Sleep(time.Second)
						_ = sub.Close()
					}),
				)
				Expect(t, err, Succeed())
				t.Log(<-rec)
			})
		})

		t.Run("HandlerPanic", func(t *testing.T) {
			ctx := hack.WithPulsar(hack.Context(t), t, dsn)
			ep := confmq.Must(ctx)

			sub, err := ep.Subscribe(ctx, genTopic(t))
			Expect(t, err, Succeed())
			Expect(t, sub.Topic(), Equal(genTopic(t)))

			rec := sub.Run(ctx, func(ctx context.Context, msg confmq.Message) error {
				panic("in consumer handler")
			})

			time.Sleep(time.Millisecond * 100)
			err = ep.Publish(ctx, confmq.NewMessage(ctx, genTopic(t), nil))
			Expect(t, err, Succeed())

			Expect(t, <-rec, ErrorContains("in consumer handler"))
		})
	})
}
