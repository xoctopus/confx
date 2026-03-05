package confpulsar_test

import (
	"testing"
	"time"

	. "cgtech.gitlab.com/saitox/x/testx"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/apache/pulsar-client-go/pulsar/backoff"

	. "cgtech.gitlab.com/saitox/confx/pkg/confpulsar"
	"cgtech.gitlab.com/saitox/confx/pkg/types/mq"
)

func TestPulsarOption(t *testing.T) {
	opt := &Option{
		EnableSubShared:    true,
		EnableRetryNack:    true,
		DisablePersistence: true,
	}
	opt.SetDefault()

	t.Run("Topic", func(t *testing.T) {
		topic := "x"
		opt.PatchTopic(&topic)
		Expect(t, topic, HavePrefix("non-persistent://"))

		opt2 := opt
		opt2.Cluster = "cluster"
		topic = "x"
		opt2.SetDefault()
		opt2.PatchTopic(&topic)
		Expect(t, topic, ContainsSubString(opt2.Cluster))
	})

	t.Run("OptionScheme", func(t *testing.T) {
		Expect(t, opt.PubOption(WithPubTopic("x")).OptionScheme(), Equal("pulsar"))
		Expect(t, opt.SubOption(WithSubTopic("x")).OptionScheme(), Equal("pulsar"))
	})

	topic := TopicFor(t)
	po := opt.PubOption(
		WithPubTopic(topic),
		WithPublishCallback(func(ProducerMessage, error) {}),
		WithSyncPublish(),
		WithPubSendTimeout(time.Minute),
		WithPubEnableBlockIfQueueFull(),
		WithPubMaxPendingMessages(101),
		WithPubEnableCompression(),
		WithPubBatchingMaxMessages(100),
		WithPubAccessMode(pulsar.ProducerAccessModeWaitForExclusive),
	).Options()
	Expect(t, po.Topic, HaveSuffix(topic))
	Expect(t, po.SendTimeout, Equal(time.Minute))
	Expect(t, po.DisableBlockIfQueueFull, BeTrue())
	Expect(t, po.MaxPendingMessages, Equal(101))
	Expect(t, po.CompressionLevel, Equal(pulsar.Default))
	Expect(t, po.CompressionType, Equal(pulsar.LZ4))
	Expect(t, po.ProducerAccessMode, Equal(pulsar.ProducerAccessModeWaitForExclusive))
	Expect(t, po.BackOffPolicyFunc(), NotBeNil[backoff.Policy]())

	po = opt.PubOption(
		WithPulsarProducerOptions(pulsar.ProducerOptions{
			Topic: topic,
			Name:  topic,
		}),
	).Options()
	Expect(t, po.Name, Equal(t.Name()))

	so := opt.SubOption(
		WithPulsarConsumerOptions(pulsar.ConsumerOptions{RetryEnable: true}),
		WithSubConsumingMode(mq.PartitionOrdered),
		WithSubDisableAutoAck(),
		WithSubCallback(func(mq.Acknowledger[ConsumerMessage], ConsumerMessage, error) {}),
		WithSubTopic(topic),
		WithSubTopic(topic, topic),
		WithSubTopicPattern("^order"),
		WithSubGroupName(topic),
		WithSubType(pulsar.Shared),
		WithSubEnableRetryNack(time.Minute, 100),
		WithSubEnableRetryNack(time.Minute, 0),
		WithSubWorkerSize(0),
		WithSubWorkerBufferSize(0),
		WithSubOrderedKeyHasher(nil),
	).Options()

	Expect(t, so.Topic, HaveSuffix(topic))
	Expect(t, so.SubscriptionName, Equal(topic))
	Expect(t, so.Type, Equal(pulsar.Shared))
	Expect(t, so.NackRedeliveryDelay, Equal(time.Minute))

	policy := so.NackBackoffPolicy
	Expect(t, policy.Next(0), Equal(time.Minute))
	Expect(t, policy.Next(2), Equal(2*time.Minute))
	Expect(t, policy.Next(100), Equal(10*time.Minute))
}
