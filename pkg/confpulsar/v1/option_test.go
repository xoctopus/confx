package confpulsar_test

import (
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/apache/pulsar-client-go/pulsar/backoff"
	. "github.com/xoctopus/x/testx"

	. "github.com/xoctopus/confx/pkg/confpulsar/v1"
	"github.com/xoctopus/confx/pkg/types/mq"
)

func TestPulsarOption(t *testing.T) {
	opt := &Option{
		EnablePubShared:   true,
		EnableSubShared:   true,
		EnableRetryNack:   true,
		DisablePersistent: true,
	}
	opt.SetDefault()

	Expect(t, opt.Topic("x"), HavePrefix("non-persistent://"))
	Expect(t, opt.PubOption().OptionScheme(), Equal("pulsar"))
	Expect(t, opt.SubOption().OptionScheme(), Equal("pulsar"))

	topic := TopicFor(t)

	po := opt.PubOption(
		WithPubTopic(topic),
		WithPublishCallback(func(mq.Message, error) {}),
		WithSyncPublish(),
		WithPubSendTimeout(time.Minute),
		WithPubEnableBlockIfQueueFull(),
		WithPubMaxPendingMessages(101),
		WithPubEnableCompression(),
		WithPubBatchingMaxMessages(100),
		WithPubAccessMode(pulsar.ProducerAccessModeWaitForExclusive),
	).Options()
	Expect(t, po.Topic, Equal(t.Name()))
	Expect(t, po.SendTimeout, Equal(time.Minute))
	Expect(t, po.DisableBlockIfQueueFull, BeTrue())
	Expect(t, po.MaxPendingMessages, Equal(101))
	Expect(t, po.CompressionLevel, Equal(pulsar.Default))
	Expect(t, po.CompressionType, Equal(pulsar.LZ4))
	Expect(t, po.ProducerAccessMode, Equal(pulsar.ProducerAccessModeWaitForExclusive))
	Expect(t, po.BackOffPolicyFunc(), Equal[backoff.Policy](&backoff.DefaultBackoff{}))

	po = opt.PubOption(
		WithPublisherOptions(pulsar.ProducerOptions{Name: topic}),
	).Options()
	Expect(t, po.Name, Equal(t.Name()))

	so := opt.SubOption(
		WithPulsarConsumerOptions(pulsar.ConsumerOptions{
			RetryEnable: true,
		}),
		WithSubDisableAutoAck(),
		WithSubCallback(func(pulsar.Consumer, pulsar.Message, mq.Message, error) {}),
		WithSubTopic(topic),
		WithSubTopic(topic, topic),
		WithSubTopicPattern("^order"),
		WithSubName(topic),
		WithSubType(pulsar.Shared),
		WithSubEnableRetryNack(time.Minute, 100),
		WithSubEnableRetryNack(time.Minute, 0),
	).Options()

	Expect(t, so.Topic, Equal(topic))
	Expect(t, so.SubscriptionName, Equal(topic))
	Expect(t, so.Type, Equal(pulsar.Shared))
	Expect(t, so.NackRedeliveryDelay, Equal(time.Minute))

	backoff := so.NackBackoffPolicy
	Expect(t, backoff.Next(0), Equal(time.Minute))
	Expect(t, backoff.Next(2), Equal(2*time.Minute))
	Expect(t, backoff.Next(100), Equal(10*time.Minute))
}
