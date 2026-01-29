package confpulsar_test

import (
	"context"
	_ "embed"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/xoctopus/x/codex"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/hack"
	"github.com/xoctopus/confx/pkg/confmq"
	"github.com/xoctopus/confx/pkg/confpulsar/v1"
	"github.com/xoctopus/confx/pkg/conftls"
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

func TopicFor(t *testing.T) string {
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

				_, err := ep.Publisher(ctx, confpulsar.WithPubTopic(TopicFor(t)))
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
			ctx, cancel := context.WithTimeout(ctx, time.Minute*2)
			defer cancel()
			ep := confmq.Must(ctx)
			Expect(t, ep, NotBeNil[confmq.PubSub]())
			topic := TopicFor(t)

			t.Run("LivenessCheck", func(t *testing.T) {
				d := ep.(types.LivenessChecker).LivenessCheck(ctx)
				Expect(t, d.Reachable, BeTrue())
			})

			pub, err := ep.Publisher(
				ctx,
				confpulsar.WithPubTopic(topic),
				confpulsar.WithPublishCallback(func(confmq.Message, error) {}),
			)
			Expect(t, err, Succeed())
			Expect(t, pub.Topic(), HaveSuffix(topic))
			Expect(t, pub.Publish(ctx, t.Name()), Succeed())

			t.Run("InvalidMessage", func(t *testing.T) {
				err = pub.PublishMessage(ctx, confmq.NewMessage(ctx, "other", nil))
				Expect(t, codex.IsCode(err, confpulsar.ECODE__PUB_INVALID_MESSAGE), BeTrue())
			})

			t.Run("PublisherClosed", func(t *testing.T) {
				pub.Close()

				err = pub.PublishMessage(ctx, confmq.NewMessage(ctx, topic, nil))
				Expect(t, codex.IsCode(err, confpulsar.ECODE__PUBLISHER_CLOSED), BeTrue())
			})

			sub, err := ep.Subscriber(
				ctx,
				confpulsar.WithSubTopic(topic),
			)
			Expect(t, err, Succeed())

			<-sub.Run(ctx, func(ctx context.Context, msg confmq.Message) error {
				Expect(t, msg.Topic(), HaveSuffix(topic))
				if string(msg.Data()) == t.Name() {
					time.Sleep(time.Second) // wait unacked messages
					sub.Close()
				}
				return nil
			})

			t.Run("EndpointClosed", func(t *testing.T) {
				Expect(t, ep.Close(), Succeed())
				err = pub.Publish(ctx, nil)
				Expect(t, codex.IsCode(err, confpulsar.ECODE__CLIENT_CLOSED), BeTrue())
			})
		})

		t.Run("FailedToInitClient", func(t *testing.T) {
			_, err := hack.TryWithPulsar(
				hack.Context(t), t,
				"pulsar://localhost:16650?connectionMaxIdleTime=30s",
			)
			expect := &pulsar.Error{}
			Expect(t, errors.As(err, &expect), BeTrue())
			Expect(t, expect.Result(), Equal(pulsar.InvalidConfiguration))
		})

		t.Run("ClosedClient", func(t *testing.T) {
			ctx := hack.WithPulsar(hack.Context(t), t, dsn)
			ep := confmq.Must(ctx)
			_ = ep.Close()

			_, err := ep.Subscriber(
				ctx,
				confpulsar.WithSubTopic(TopicFor(t)),
			)
			Expect(t, err, Failed())
			_, err = ep.Publisher(ctx, confpulsar.WithPubTopic(TopicFor(t)))
			Expect(t, err, Failed())
			r := ep.(types.LivenessChecker).LivenessCheck(ctx)
			Expect(t, r.Reachable, BeFalse())
		})

		t.Run("SendMode", func(t *testing.T) {
			t.Run("Sync", func(t *testing.T) {
				var (
					ctx = hack.WithPulsar(hack.Context(t), t, dsn)
					ep  = confmq.Must(ctx)
				)
				sub, err := ep.Subscriber(ctx, confpulsar.WithSubTopic(TopicFor(t)))
				Expect(t, err, Succeed())
				pub, err := ep.Publisher(ctx, confpulsar.WithPubTopic(TopicFor(t)))
				Expect(t, err, Succeed())

				err = pub.Publish(ctx, "send_mode:sync")
				Expect(t, err, Succeed())

				<-sub.Run(
					ctx, func(ctx context.Context, msg confmq.Message) error {
						raw := string(msg.Data())
						Expect(t, raw, Equal("send_mode:sync"))
						time.Sleep(time.Second)
						sub.Close()
						return nil
					},
				)
			})
			t.Run("Async", func(t *testing.T) {
				var (
					ctx = hack.WithPulsar(hack.Context(t), t, dsn)
					ep  = confmq.Must(ctx)
				)
				sub, err := ep.Subscriber(ctx, confpulsar.WithSubTopic(TopicFor(t)))
				Expect(t, err, Succeed())
				pub, err := ep.Publisher(ctx, confpulsar.WithPubTopic(TopicFor(t)))
				Expect(t, err, Succeed())

				err = pub.Publish(ctx, "send_mode:async")
				Expect(t, err, Succeed())

				<-sub.Run(ctx, func(ctx context.Context, msg confmq.Message) error {
					raw := string(msg.Data())
					Expect(t, raw, Equal("send_mode:async"))
					time.Sleep(time.Second)
					sub.Close()
					return nil
				})
			})
		})

		t.Run("HandlerPanic", func(t *testing.T) {
			var (
				ctx = hack.WithPulsar(hack.Context(t), t, dsn)
				ep  = confmq.Must(ctx)
			)
			sub, err := ep.Subscriber(ctx,
				confpulsar.WithSubTopic(TopicFor(t)),
				confpulsar.WithSubDisableAutoAck(),
				confpulsar.WithSubCallback(func(c pulsar.Consumer, m pulsar.Message, p confmq.Message, err error) {
					_ = c.Ack(m)
					if err == nil {
						return
					}
					if codex.IsCode(err, confpulsar.ECODE__PARSE_MESSAGE) ||
						codex.IsCode(err, confpulsar.ECODE__HANDLER_PANICKED) {
						Expect(t, err, ErrorContains("in consumer handler"))
					}
				}),
			)
			Expect(t, err, Succeed())
			pub, err := ep.Publisher(ctx, confpulsar.WithPubTopic(TopicFor(t)))
			Expect(t, err, Succeed())

			err = pub.Publish(ctx, "any")
			Expect(t, err, Succeed())

			<-sub.Run(ctx, func(ctx context.Context, msg confmq.Message) error {
				defer sub.Close()
				panic("in consumer handler")
			})
		})

		t.Run("MultiLivenessCheck", func(t *testing.T) {
			var (
				err1 error
				err2 error
				wg   sync.WaitGroup
			)

			wg.Add(2)
			t.Run("Checker1", func(t *testing.T) {
				go func() {
					defer wg.Done()
					ctx := hack.Context(t)
					_, err1 = hack.TryWithPulsar(ctx, t, dsn)
					t.Log(err1)
				}()
			})
			t.Run("Checker2", func(t *testing.T) {
				go func() {
					defer wg.Done()
					ctx := hack.Context(t)
					_, err1 = hack.TryWithPulsar(ctx, t, dsn)
					t.Log(err2)
				}()
			})
			wg.Wait()

			// expect at-least-one success
			Expect(t, err1 == nil || err2 == nil, BeTrue())
		})
	})
}
