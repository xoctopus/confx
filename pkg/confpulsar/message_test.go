package confpulsar_test

import (
	"testing"
	"time"

	. "github.com/xoctopus/x/testx"

	. "github.com/xoctopus/confx/pkg/confpulsar"
)

func TestProducerMessage(t *testing.T) {
	var (
		topic = TopicFor(t)
		body  = []byte(topic)
	)

	mp := NewProducerMessage(topic, body)
	mp.SetPartitionKey("abc")
	mp.SetOrderingKey("ordering_key")
	mp.SetDelay(time.Second)

	Expect(t, mp.Topic(), Equal(topic))
	Expect(t, mp.Payload(), Equal(body))
	Expect(t, mp.PartitionKey(), Equal("abc"))
	Expect(t, mp.OrderingKey(), Equal("ordering_key"))
	Expect(t, time.Since(mp.PublishedAt()) > 0, BeTrue())
	Expect(t, mp.Delay(), Equal(time.Second))

	Expect(t, len(mp.Extra()), Equal(1))
	v, _ := mp.ExtraValueOf(EXTRA_KIND__DELIVERY_DELAYED.String())
	Expect(t, v, Equal("1000"))

	// modify through underlying
	u := mp.Underlying()
	u.Key = "partition_key"
	Expect(t, mp.PartitionKey(), Equal("partition_key"))
}
