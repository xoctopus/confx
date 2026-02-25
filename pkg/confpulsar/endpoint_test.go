package confpulsar_test

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/google/uuid"
	. "github.com/xoctopus/x/testx"
	"github.com/xoctopus/x/testx/bdd"

	"github.com/xoctopus/confx/hack"
	. "github.com/xoctopus/confx/pkg/confpulsar"
	"github.com/xoctopus/confx/pkg/testdata"
	"github.com/xoctopus/confx/pkg/types"
	"github.com/xoctopus/confx/pkg/types/liveness"
	"github.com/xoctopus/confx/pkg/types/mq"
)

func TopicFor(t testing.TB) string {
	return strings.ReplaceAll(t.Name(), "/", "_")
}

func TestEndpoint(t *testing.T) {
	bdd.From(t).Given("EmptyEndpoint", func(t bdd.T) {
		ep := &Endpoint{}
		t.When("SetDefault", func(t bdd.T) {
			ep.SetDefault()
			t.Then(
				"AddressIsDefaultLocalPulsarEndpoint",
				bdd.Equal(ep.Address, "pulsar://localhost:6650"),
			)
		})
	})

	bdd.From(t).Given("InvalidDSN", func(t bdd.T) {
		ep := &Endpoint{
			Endpoint: types.Endpoint[Option]{
				Address: "pulsar://localhost:6379/%zz",
			},
		}
		t.When("Init", func(b bdd.T) {
			err := ep.Init(hack.Context(t))
			_, ok := errors.AsType[*url.Error](err)
			t.Then("FailedToInitCausedByURL", bdd.BeTrue(ok))
		})
	})

	bdd.From(t).Given("UnreachableDSNAndTLSConfig", func(t bdd.T) {
		ep := &Endpoint{
			Endpoint: types.Endpoint[Option]{
				Address: "pulsar://localhost:6650",
				Cert:    testdata.TLSConfigForTest(),
			},
		}
		t.When("InitWithTLS", func(t bdd.T) {
			ep.SetDefault()
			err := ep.Init(hack.Context(t))
			t.Then("FailedToInitCausedByTimeout", bdd.ErrorContains("", err))
		})
	})

	bdd.From(t).Given("UnreachableDSN", func(t bdd.T) {
		var dsn = "pulsar://localhost:6650?connectionTimeout=100ms"
		t.When("TryToConnect", func(t bdd.T) {
			ctx := hack.WithPulsarLost(hack.Context(t), t, dsn)
			t.When("CheckLiveness", func(t bdd.T) {
				err := Must(ctx).(*Endpoint).LivenessCheck(ctx).FailureReason()
				t.Then("Failed", bdd.ErrorContains("connection error", err))
			})
		})
	})

	bdd.From(t).Given("InvalidEndpointOption", func(t bdd.T) {
		dsn := "pulsar://localhost:16650?connectionMaxIdleTime=30s"
		t.When("TryToConnect", func(t bdd.T) {
			ep := Endpoint{}
			ep.Address = dsn
			ep.SetDefault()
			err := ep.Init(hack.Context(t))
			t.Then(
				"FailedToInitCausedByInvalidPulsarConfig",
				bdd.IsCodeError(ERROR__CLI_INIT_ERROR, err),
			)
			_ = ep.Close()
		})
	})

	dsn := "pulsar://localhost:16650"
	bdd.From(t).Given("ReachableDSN", func(t bdd.T) {
		var (
			ctx = hack.WithPulsar(hack.Context(t), t, dsn)
			ps  = Must(ctx)
		)
		t.Then("Reachable", bdd.NotBeNil(ps))
	})

	t.Run("ValidateConsumerMessage", func(t *testing.T) {
		var (
			mp      = NewProducerMessage(TopicFor(t), []byte(uuid.NewString()))
			termsig = make(chan struct{}, 1)
		)
		mp.SetDelay(time.Second)
		mp.SetExpiredAt(9000000000)

		hack.RunPulsarPubSubTestSuite(
			hack.Context(t), t, dsn,
			[]ProducerMessage{mp},
			termsig,
			func(ctx context.Context, mc ConsumerMessage) error {
				if !bytes.Equal(mp.Payload(), mc.Payload()) {
					return nil
				}
				Expect(t, mc.Topic(), HaveSuffix(mp.Topic()))
				Expect(t, mc.PartitionKey(), Equal(mp.PartitionKey()))
				Expect(t, mc.OrderingKey(), Equal(mp.OrderingKey()))
				Expect(t, mc.PublishedAt().Second(), Equal(mp.PublishedAt().Second()))
				Expect(t, mc.RetryCount(), Equal(uint32(0)))
				Expect(t, mc.ProducedBy(), Equal(mp.Topic()))
				Expect(t, mc.Underlying().ID().PartitionIdx(), Equal(int32(mc.PartitionID())))

				v, _ := mc.ExtraValueOf(EXTRA_KEY__EXPIRED_AT)
				Expect(t, v, Equal("9000000000"))
				_, ok := mc.Extra()["none"]
				Expect(t, ok, BeFalse())
				_, ok = mc.ExtraValueOf("none")
				Expect(t, ok, BeFalse())

				d1 := mc.Latency()
				d2 := mc.BrokerLatency()
				Expect(t, d2 == 0 || d1 < d2, BeTrue())

				termsig <- struct{}{}
				return nil
			},
			time.Second*5,
			true,
			WithPubTopic(mp.Topic()),
			WithPublisherName(mp.Topic()),
			WithSubTopic(mp.Topic()),
		)
	})

	t.Run("Producing", func(t *testing.T) {
		var (
			topic = TopicFor(t)
			ctx   = hack.WithPulsar(hack.Context(t), t, dsn)
		)

		t.Run("InvalidProducerOption", func(t *testing.T) {
			_, err := Must(ctx).NewProducer(
				ctx,
				WithPulsarProducerOptions(pulsar.ProducerOptions{
					DisableBatching: false,
					EnableChunking:  true,
				}),
				WithPubTopic(topic),
				WithPubAccessMode(pulsar.ProducerAccessModeShared),
			)
			Expect(t, err, Failed())
		})
		t.Run("InvalidTopic", func(t *testing.T) {
			pub, err := Must(ctx).NewProducer(
				ctx,
				WithPubTopic(topic),
				WithPubAccessMode(pulsar.ProducerAccessModeShared),
			)
			Expect(t, pub.Topic(), HaveSuffix(topic))
			Expect(t, err, Succeed())
			_, err = pub.Publish(ctx, "unmatched", nil)
			Expect(t, err, IsCodeError(ERROR__PUB_INVALID_MESSAGE))
		})
		t.Run("PubAsyncWithCallback", func(t *testing.T) {
			pub, err := Must(ctx).NewProducer(
				ctx,
				WithPubTopic(topic),
				WithPubAccessMode(pulsar.ProducerAccessModeShared),
				WithPublishCallback(func(_ ProducerMessage, err error) {
					Expect(t, err, Succeed())
				}),
			)
			Expect(t, err, Succeed())
			_, err = pub.Publish(ctx, topic, nil)
			time.Sleep(time.Second << 1)
		})
		t.Run("PublisherClosed", func(t *testing.T) {
			pub, err := Must(ctx).NewProducer(
				ctx,
				WithPubTopic(topic),
				WithPubAccessMode(pulsar.ProducerAccessModeShared),
			)
			Expect(t, err, Succeed())
			Expect(t, pub.Close(), Succeed())
			_, err = pub.PublishWithKey(ctx, topic, "partition-key", nil)
			Expect(t, err, IsCodeError(ERROR__PUB_CLOSED))
		})
		t.Run("ClientClosed", func(t *testing.T) {
			pub, err := Must(ctx).NewProducer(
				ctx,
				WithPubTopic(topic),
				WithPubAccessMode(pulsar.ProducerAccessModeShared),
			)
			Expect(t, err, Succeed())
			Expect(t, Must(ctx).Close(), Succeed())
			_, err = Must(ctx).NewProducer(ctx, WithPubTopic(topic))
			Expect(t, err, IsCodeError(ERROR__CLI_CLOSED))
			err = pub.PublishMessage(ctx, NewProducerMessage(topic, nil))
			Expect(t, err, IsCodeError(ERROR__CLI_CLOSED))
		})
	})

	t.Run("Consuming", func(t *testing.T) {
		var (
			topic = TopicFor(t)
			ctx   = hack.WithPulsar(hack.Context(t), t, dsn)
			hdl   = func(context.Context, ConsumerMessage) error { return nil }
		)

		t.Run("InvalidConsumerOption", func(t *testing.T) {
			_, err := Must(ctx).NewConsumer(
				ctx,
				WithPulsarConsumerOptions(pulsar.ConsumerOptions{
					RetryEnable:             true,
					EnableZeroQueueConsumer: true,
				}),
				WithSubTopic(topic),
			)
			Expect(t, err, Failed())
		})
		t.Run("Rerun", func(t *testing.T) {
			sub, err := Must(ctx).NewConsumer(ctx, WithSubTopic(topic), WithSubWorkerSize(2), WithSubWorkerBufferSize(1))
			Expect(t, err, Succeed())
			go func() {
				_ = sub.Run(ctx, hdl)
			}()
			time.Sleep(time.Second)
			err = sub.Run(ctx, hdl)
			Expect(t, err, IsCodeError(ERROR__SUB_BOOTED))
		})
		t.Run("ConsumerClosed", func(t *testing.T) {
			sub, err := Must(ctx).NewConsumer(ctx, WithSubTopic(topic))
			Expect(t, err, Succeed())
			Expect(t, sub.Close(), Succeed())
			Expect(t, sub.Run(ctx, hdl), IsCodeError(ERROR__SUB_CLOSED))
		})
		t.Run("ClientClosed", func(t *testing.T) {
			sub, err := Must(ctx).NewConsumer(ctx, WithSubTopic(topic))
			Expect(t, err, Succeed())
			Expect(t, Must(ctx).Close(), Succeed())

			_, err = Must(ctx).NewConsumer(ctx, WithSubTopic(topic))
			Expect(t, err, IsCodeError(ERROR__CLI_CLOSED))

			v := Must(ctx).(liveness.Checker).LivenessCheck(ctx)
			Expect(t, v.FailureReason(), IsCodeError(ERROR__CLI_CLOSED))

			Expect(t, sub.Run(ctx, hdl), IsCodeError(ERROR__CLI_CLOSED))
		})
	})

	t.Run("ConsumerHandling", func(t *testing.T) {
		var herr = errors.New("handle error")
		t.Run("HandlerPanicked", func(t *testing.T) {
			var (
				topic   = TopicFor(t)
				msg     = NewProducerMessage(topic, []byte(uuid.NewString()))
				termsig = make(chan struct{}, 1)
				handler = func(_ context.Context, m ConsumerMessage) error {
					if bytes.Equal(m.Payload(), msg.Payload()) {
						termsig <- struct{}{}
						panic(herr)
					}
					return nil
				}
			)

			hack.RunPulsarPubSubTestSuite(
				hack.Context(t), t, dsn,
				[]ProducerMessage{msg},
				termsig,
				handler,
				time.Second*5,
				false,
				WithPubTopic(topic),
				WithSubTopic(topic),
				WithSubConsumingMode(mq.PartitionOrdered),
				WithSubWorkerSize(1),
				WithSubWorkerBufferSize(1),
				WithSubCallback(func(_ mq.Acknowledger[ConsumerMessage], m ConsumerMessage, err error) {
					if bytes.Equal(m.Payload(), msg.Payload()) {
						Expect(t, err, IsError(herr))
						Expect(t, err, IsCodeError(ERROR__SUB_HANDLER_PANICKED))
					}
				}),
			)
		})
		t.Run("HandlerFailed", func(t *testing.T) {
			var (
				topic   = TopicFor(t)
				msg     = NewProducerMessage(topic, []byte(uuid.NewString()))
				termsig = make(chan struct{}, 1)
				handler = func(_ context.Context, m ConsumerMessage) error {
					if bytes.Equal(m.Payload(), msg.Payload()) {
						return herr
					}
					return nil
				}
			)

			hack.RunPulsarPubSubTestSuite(
				hack.Context(t), t, dsn,
				[]ProducerMessage{msg},
				termsig,
				handler,
				time.Second*5,
				false,
				WithPubTopic(topic),
				WithSubTopic(topic),
				WithSubConsumingMode(mq.Concurrent),
				WithSubWorkerSize(1),
				WithSubWorkerBufferSize(1),
				WithSubDisableAutoAck(),
				WithSubCallback(func(a mq.Acknowledger[ConsumerMessage], m ConsumerMessage, err error) {
					if bytes.Equal(m.Payload(), msg.Payload()) {
						Expect(t, err, IsError(herr))
					}
					if m.RetryCount() > 1 {
						_ = a.Ack(m)
						termsig <- struct{}{}
					} else {
						_ = a.Nack(m)
					}
				}),
			)
		})
	})
}

// FOR testing gracefully closing and side effects when closing rudely
/*
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
	testx.Expect(t, err, testx.Succeed())

	for range 1000 {
		err = pub.Publish(ctx, nil)
		testx.Expect(t, err, testx.Succeed())
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
	testx.Expect(t, err, testx.Succeed())

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
