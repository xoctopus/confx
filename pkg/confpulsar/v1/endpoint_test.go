package confpulsar_test

import (
	"context"
	_ "embed"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/xoctopus/x/testx/bdd"

	"github.com/xoctopus/confx/hack"
	"github.com/xoctopus/confx/pkg/confpulsar/v1"
	"github.com/xoctopus/confx/pkg/conftls"
	"github.com/xoctopus/confx/pkg/types"
	"github.com/xoctopus/confx/pkg/types/mq"
)

var (
	//go:embed testdata/client.key
	key string
	//go:embed testdata/client.crt
	crt string
	//go:embed testdata/ca.crt
	ca string
)

func TopicFor(t testing.TB) string {
	return strings.ReplaceAll(t.Name(), "/", "_")
}

/*
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
				ep := mq.Must(ctx)
				Expect(t, ep, NotBeNil[mq.PubSub]())

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
			ep := mq.Must(ctx)
			Expect(t, ep, NotBeNil[mq.PubSub]())
			topic := TopicFor(t)

			t.Run("LivenessCheck", func(t *testing.T) {
				d := ep.(types.LivenessChecker).LivenessCheck(ctx)
				Expect(t, d.Reachable, BeTrue())
			})

			pub, err := ep.Publisher(
				ctx,
				confpulsar.WithPubTopic(topic),
				confpulsar.WithPublishCallback(func(mq.Message, error) {}),
			)
			Expect(t, err, Succeed())
			Expect(t, pub.Topic(), HaveSuffix(topic))
			Expect(t, pub.Publish(ctx, t.Name()), Succeed())

			t.Run("InvalidMessage", func(t *testing.T) {
				err = pub.PublishMessage(ctx, mq.NewMessage(ctx, "other", nil))
				Expect(t, codex.IsCode(err, confpulsar.ECODE__PUB_INVALID_MESSAGE), BeTrue())
			})

			t.Run("PublisherClosed", func(t *testing.T) {
				pub.Close()

				err = pub.PublishMessage(ctx, mq.NewMessage(ctx, topic, nil))
				Expect(t, codex.IsCode(err, confpulsar.ECODE__PUBLISHER_CLOSED), BeTrue())
			})

			sub, err := ep.Subscriber(
				ctx,
				confpulsar.WithSubTopic(topic),
			)
			Expect(t, err, Succeed())

			<-sub.Run(ctx, func(ctx context.Context, msg mq.Message) error {
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
			ep := mq.Must(ctx)
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
					ep  = mq.Must(ctx)
				)
				sub, err := ep.Subscriber(ctx, confpulsar.WithSubTopic(TopicFor(t)))
				Expect(t, err, Succeed())
				pub, err := ep.Publisher(ctx, confpulsar.WithPubTopic(TopicFor(t)))
				Expect(t, err, Succeed())

				err = pub.Publish(ctx, "send_mode:sync")
				Expect(t, err, Succeed())

				<-sub.Run(
					ctx, func(ctx context.Context, msg mq.Message) error {
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
					ep  = mq.Must(ctx)
				)
				sub, err := ep.Subscriber(ctx, confpulsar.WithSubTopic(TopicFor(t)))
				Expect(t, err, Succeed())
				pub, err := ep.Publisher(ctx, confpulsar.WithPubTopic(TopicFor(t)))
				Expect(t, err, Succeed())

				err = pub.Publish(ctx, "send_mode:async")
				Expect(t, err, Succeed())

				<-sub.Run(ctx, func(ctx context.Context, msg mq.Message) error {
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
				ep  = mq.Must(ctx)
			)
			sub, err := ep.Subscriber(ctx,
				confpulsar.WithSubTopic(TopicFor(t)),
				confpulsar.WithSubDisableAutoAck(),
				confpulsar.WithSubCallback(func(c pulsar.Consumer, m pulsar.Message, p mq.Message, err error) {
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

			<-sub.Run(ctx, func(ctx context.Context, msg mq.Message) error {
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
*/

func TestEndpointV1_2(t *testing.T) {
	bdd.From(t).Given("EmptyEndpoint", func(t bdd.T) {
		ep := &confpulsar.Endpoint{}
		t.When("SetDefault", func(t bdd.T) {
			ep.SetDefault()
			t.Then(
				"AddressIsDefaultLocalPulsarEndpoint",
				bdd.Equal(ep.Address, "pulsar://localhost:6650"),
			)
		})
	})

	bdd.From(t).Given("InvalidDSN", func(t bdd.T) {
		ep := &confpulsar.Endpoint{
			Endpoint: types.Endpoint[confpulsar.Option]{
				Address: "pulsar://localhost:6379/%zz",
			},
		}
		t.When("Init", func(b bdd.T) {
			err := ep.Init(hack.Context(t))
			err2 := &url.Error{}
			t.Then("FailedToInitCausedByURL", bdd.BeTrue(errors.As(err, &err2)))
		})
	})

	bdd.From(t).Given("UnreachableDSNAndTLSConfig", func(t bdd.T) {
		ep := &confpulsar.Endpoint{
			Endpoint: types.Endpoint[confpulsar.Option]{
				Address: "pulsar://localhost:6650",
				Cert:    conftls.X509KeyPair{Key: key, Crt: crt, CA: ca},
			},
		}
		t.When("InitWithTLS", func(t bdd.T) {
			ep.SetDefault()
			err := ep.Init(hack.Context(t))
			t.Then("FailedToInitCausedByTimeout", bdd.ErrorContains("", err))
		})
	})

	bdd.From(t).Given("UnreachableDSN", func(t bdd.T) {
		dsn := "pulsar://localhost:6650?connectionTimeout=100ms"
		t.When("TryToConnectAndNewProducer", func(t bdd.T) {
			ctx := hack.WithPulsarLost(hack.Context(t), t, dsn)
			_, err := mq.Must(ctx).Publisher(ctx, confpulsar.WithPubTopic(TopicFor(t)))
			t.Then("GotConnectionError", bdd.ErrorContains("connection error", err))
		})
	})

	bdd.From(t).Given("InvalidEndpointOption", func(t bdd.T) {
		dsn := "pulsar://localhost:16650?connectionMaxIdleTime=30s"
		t.When("Init", func(t bdd.T) {
			ep := confpulsar.Endpoint{
				Endpoint: types.Endpoint[confpulsar.Option]{
					Address: dsn,
				},
			}
			defer func() { _ = ep.Close() }()
			ep.SetDefault()
			err := ep.Init(hack.Context(t))
			t.Then(
				"FailedToInitCausedByInvalidPulsarConfig",
				bdd.IsCodeError(confpulsar.ECODE__FAILED_TO_INIT_CLIENT, err),
			)
		})
	})

	bdd.From(t).Given("ReachableEndpoint", func(t bdd.T) {
		dsn := "pulsar://localhost:16650"
		t.When("Init", func(t bdd.T) {
			var (
				ctx    = hack.WithPulsar(hack.Context(t), t, dsn)
				ps     = mq.Must(ctx)
				cause  = errors.New("exceeded unit test deadline")
				cancel context.CancelFunc
			)
			// force finished after 1 minute
			ctx, cancel = context.WithTimeoutCause(ctx, time.Minute, cause)

			defer func() {
				_ = ps.Close()
				cancel()
			}()

			t.Then("Reachable", bdd.NotBeNil(ps))

			var (
				pub   mq.Publisher
				sub   mq.Subscriber
				err   error
				msg   mq.Message
				topic = "TestPulsarEndpoint"
			)

			t.When("CreatePublisher", func(t bdd.T) {
				pub, err = ps.Publisher(
					ctx,
					confpulsar.WithPubTopic(topic),
					confpulsar.WithSyncPublish(),
				)
				t.Then("Succeed", bdd.Succeed(err))
			})

			t.When("CreateSubscriber", func(t bdd.T) {
				sub, err = ps.Subscriber(
					ctx,
					confpulsar.WithSubTopic(topic),
					confpulsar.WithSubConsumingMode(mq.PartitionOrdered),
					confpulsar.WithSubWorkerSize(8),
					confpulsar.WithSubWorkerBufferSize(4),
				)
				t.Then("Succeed", bdd.Succeed(err))
			})

			t.When("CreatePublishersAndSubscribers", func(t bdd.T) {
				_, err = ps.Publisher(ctx, confpulsar.WithPubTopic("1"))
				t.Then("Succeed", bdd.Succeed(err))
				_, err = ps.Publisher(ctx, confpulsar.WithPubTopic("2"))
				t.Then("Succeed", bdd.Succeed(err))
				_, err = ps.Publisher(ctx, confpulsar.WithPubTopic("3"))
				t.Then("Succeed", bdd.Succeed(err))

				_, err = ps.Subscriber(ctx, confpulsar.WithSubTopic("1"))
				t.Then("Succeed", bdd.Succeed(err))
				_, err = ps.Subscriber(ctx, confpulsar.WithSubTopic("2"))
				t.Then("Succeed", bdd.Succeed(err))
				_, err = ps.Subscriber(ctx, confpulsar.WithSubTopic("3"))
				t.Then("Succeed", bdd.Succeed(err))

				ep := ps.(*confpulsar.Endpoint)
				t.Then("MatchConsumerCount", bdd.Equal(ep.ConsumerCount(), 4))
				t.Then("MatchProducerCount", bdd.Equal(ep.ProducerCount(), 4))
			})

			t.When("PublishMessageNotMatchPublisherTopic", func(t bdd.T) {
				err = pub.PublishMessage(ctx, mq.NewMessage(ctx, "any_other", nil))
				t.Then(
					"FailedCausedByInvalidMessage",
					bdd.IsCodeError(confpulsar.ECODE__PUB_INVALID_MESSAGE, err),
				)
			})

			t.When("PublishMessage", func(t bdd.T) {
				msg = mq.NewMessage(ctx, topic, "any")
				msg.(mq.MutMessage).SetPubOrderedKey("biz_ordered_key")

				err = pub.PublishMessage(ctx, msg)
				t.Then("Success", bdd.Succeed(err))
			})

			t.When("PublishOnClosedPublisher", func(t bdd.T) {
				pub.Close()
				err = pub.PublishMessage(ctx, mq.NewMessage(ctx, topic, nil))
				t.Then(
					"FailedCausedByPublisherClosedError",
					bdd.IsCodeError(confpulsar.ECODE__PUBLISHER_CLOSED, err),
				)
			})

			t.When("Subscribing", func(t bdd.T) {
				var msgID int64
				err = sub.Run(ctx, func(ctx context.Context, m mq.Message) error {
					if m.ID() == msg.ID() {
						msgID = msg.ID()
						sub.Close()
					}
					return nil
				})
				t.Then(
					"SubscriptionCanceledAfterSubscribed",
					bdd.IsCodeError(confpulsar.ECODE__SUBSCRIPTION_CANCELED, err),
				)
				t.Then("ExpectedMessage", bdd.Equal(msgID, msg.ID()))
			})

			t.When("RerunSubscribing", func(t bdd.T) {
				err = sub.Run(ctx, nil)
				t.Then(
					"FailedCausedByRerunError",
					bdd.IsCodeError(confpulsar.ECODE__SUBSCRIBER_BOOTED, err),
				)
			})

			t.When("EndpointClosed", func(t bdd.T) {
				t.Then("ClosedSucceed", bdd.Succeed(ps.Close()))

				t.When("CreatePublisher", func(t bdd.T) {
					_, err = ps.Publisher(ctx, confpulsar.WithPubTopic("any"))
					t.Then(
						"FailedCausedByClientClosed",
						bdd.IsCodeError(confpulsar.ECODE__CLIENT_CLOSED, err),
					)
				})
				t.When("CreateSubscriber", func(t bdd.T) {
					_, err = ps.Subscriber(ctx, confpulsar.WithPubTopic("any"))
					t.Then(
						"FailedCausedByClientClosed",
						bdd.IsCodeError(confpulsar.ECODE__CLIENT_CLOSED, err),
					)
				})
				t.When("LivenessCheck", func(t bdd.T) {
					r := ps.(types.LivenessChecker).LivenessCheck(ctx)
					t.Then(
						"FailedCausedByClientClosed",
						bdd.ContainsSubString(r.Message, "client closed"),
					)
				})

				t.When("Publish", func(t bdd.T) {
					err = pub.Publish(ctx, nil)
					t.Then(
						"FailedCausedByClientClosedError",
						bdd.IsCodeError(confpulsar.ECODE__CLIENT_CLOSED, err),
					)
				})

				ep := ps.(*confpulsar.Endpoint)
				t.Then("MatchProducerCount", bdd.BeTrue(ep.ConsumerCount() == 0))
				t.Then("MatchConsumerCount", bdd.BeTrue(ep.ProducerCount() == 0))
			})
		})
	})
}

/* FOR testing gracefully closing and side effects when closing rudely

func TestPublishMany(t *testing.T) {
	var (
		ctx = hack.WithPulsar(hack.Context(t), t, "pulsar://localhost:16650")
		cli = mq.Must(ctx)
	)

	pub, err := cli.Publisher(
		ctx,
		confpulsar.WithPubTopic("TEST_RUDE_CLOSE"),
		confpulsar.WithSyncPublish(),
	)
	Expect(t, err, Succeed())

	for range 1000 {
		err = pub.Publish(ctx, nil)
		Expect(t, err, Succeed())
	}
}

func TestSubscribe_NoClose(t *testing.T) {
	var (
		ctx    = hack.WithPulsar(hack.Context(t), t, "pulsar://localhost:16650")
		cli    = mq.Must(ctx)
		cancel context.CancelFunc
	)

	once := 0

	sub, err := cli.Subscriber(
		ctx,
		confpulsar.WithSubTopic("TEST_RUDE_CLOSE"),
		confpulsar.WithSubWorkerSize(1),
		confpulsar.WithSubWorkerBufferSize(1),
		confpulsar.WithSubDisableAutoAck(),
		confpulsar.WithSubCallback(func(s pulsar.Consumer, m pulsar.Message, _ mq.Message, err error) {
			if once == 0 {
				_ = s.Ack(m)
			}
			once++
		}),
	)
	Expect(t, err, Succeed())

	ctx, cancel = context.WithTimeout(ctx, time.Second<<1)
	cancel()

	err = sub.Run(ctx, func(ctx context.Context, msg mq.Message) error {
		t.Log(msg.ID())
		return nil
	})
	t.Log(err)
	sub.Close()
	time.Sleep(time.Hour)
}
*/
