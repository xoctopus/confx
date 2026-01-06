package confpulsar

import (
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/components/confmq"
	"github.com/xoctopus/confx/pkg/types"
)

func Test_newPubOption(t *testing.T) {
	global :=
		&PulsarOption{
			OperationTimeout:  types.Duration(100 * time.Millisecond),
			ConnTimeout:       types.Duration(3 * time.Second),
			KeepAliveInterval: types.Duration(time.Hour),
			MaxConnector:      10,
			MaxPending:        100,
			MaxBatching:       100,
		}

	o := newPubOption(
		global,
		"test",
		WithPublishCallback(func(message confmq.Message, err error) {}),
		WithPublishSync(),
		WithPublishTimeout(time.Hour),
		WithMaxReconnectToBroker(1),
		WithProducerName("test"),
	)
	Expect(t, o.Scheme(), Equal("pulsar"))
	Expect(t, o.callback != nil, BeTrue())
	Expect(t, o.sync, BeTrue())
	Expect(t, o.options.SendTimeout, Equal(time.Hour))
	Expect(t, *o.options.MaxReconnectToBroker, Equal(uint(1)))
	Expect(t, o.options.Name, Equal("test"))
	Expect(t, o.options.MaxPendingMessages, Equal(100))
	Expect(t, o.options.BatchingMaxMessages, Equal(uint(100)))

	o = newPubOption(
		global, "test",
		WithPulsarProducerOptions(pulsar.ProducerOptions{
			Topic:               "test2",
			Name:                "test2",
			SendTimeout:         time.Second,
			MaxPendingMessages:  101,
			BatchingMaxMessages: 102,
		}),
	)

	Expect(t, o.Scheme(), Equal("pulsar"))
	Expect(t, o.options.Topic, Equal("test"))
	Expect(t, o.options.Name, Equal("test2"))
	Expect(t, o.options.SendTimeout, Equal(time.Second))
	Expect(t, o.options.MaxPendingMessages, Equal(101))
	Expect(t, o.options.BatchingMaxMessages, Equal(uint(102)))
}

func Test_newSubOption(t *testing.T) {
	o := newSubOption(
		&PulsarOption{}, "test",
		WithSubTopicsPattern("^order*"),
		WithSubTopics("a", "b", "c"),
		WithSubName("bad"),
		WithPulsarConsumerOptions(pulsar.ConsumerOptions{Name: "good"}),
	)

	Expect(t, o.Scheme(), Equal("pulsar"))
	Expect(t, o.options.Topic, Equal("test"))
	Expect(t, o.options.Topics, BeNil[[]string]())
	Expect(t, o.options.TopicsPattern, Equal(""))
	Expect(t, o.options.Name, Equal("good"))
}
