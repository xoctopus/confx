package confpulsar_test

import (
	"strconv"
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

	exp := time.Now().Unix() + 100
	mp.SetExpiredAt(exp)

	Expect(t, mp.Topic(), Equal(topic))
	Expect(t, mp.Payload(), Equal(body))
	Expect(t, mp.PartitionKey(), Equal("abc"))
	Expect(t, mp.OrderingKey(), Equal("ordering_key"))
	Expect(t, time.Since(mp.PublishedAt()) > 0, BeTrue())
	Expect(t, mp.Delay(), Equal(time.Second))

	Expect(t, len(mp.Extra()), Equal(1))
	v, _ := mp.ExtraValueOf(EXTRA_KEY__EXPIRED_AT)
	Expect(t, v, Equal(strconv.FormatInt(exp, 10)))

	// modify through underlying
	u := mp.Underlying()
	u.Key = "partition_key"
	Expect(t, mp.PartitionKey(), Equal("partition_key"))
}
