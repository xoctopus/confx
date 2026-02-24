package confpulsar

import (
	"strconv"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"

	"github.com/xoctopus/confx/pkg/types/mq"
)

type ProducerMessage interface {
	mq.HasTopic
	mq.CanSetTopic
	mq.HasPayload
	mq.CanSetPayload
	mq.HasExtra
	mq.CanAppendExtra
	mq.HasPartitionKey
	mq.CanSetPartitionKey
	mq.HasOrderingKey
	mq.CanSetOrderingKey
	mq.HasPublishedAt
	mq.CanRefreshPublishedAt
	mq.HasDelay
	mq.CanSetDelay
	mq.HasUnderlying[*pulsar.ProducerMessage]
}

func NewProducerMessage(topic string, payload []byte) ProducerMessage {
	m := &producerMessage{}
	m.SetTopic(topic)
	m.SetPayload(payload)
	m.RefreshPublishedAt()
	return m
}

type producerMessage struct {
	topic string
	pulsar.ProducerMessage
}

func (x *producerMessage) Topic() string {
	return x.topic
}

func (x *producerMessage) SetTopic(topic string) {
	x.topic = topic
}

func (x *producerMessage) Payload() []byte {
	return x.ProducerMessage.Payload
}

func (x *producerMessage) SetPayload(payload []byte) {
	x.ProducerMessage.Payload = payload
}

func (x *producerMessage) Extra() map[string]string {
	return x.Properties
}

func (x *producerMessage) ExtraValueOf(k string) (string, bool) {
	v, ok := x.Properties[k]
	return v, ok
}

func (x *producerMessage) AddExtra(k, v string) {
	if x.Properties == nil {
		x.Properties = make(map[string]string)
	}
	x.Properties[k] = v
}

func (x *producerMessage) PartitionKey() string {
	return x.Key
}

func (x *producerMessage) SetPartitionKey(k string) {
	x.Key = k
}

func (x *producerMessage) OrderingKey() string {
	return x.ProducerMessage.OrderingKey
}

func (x *producerMessage) SetOrderingKey(k string) {
	x.ProducerMessage.OrderingKey = k
}

func (x *producerMessage) PublishedAt() time.Time {
	return x.EventTime
}

func (x *producerMessage) RefreshPublishedAt() {
	x.EventTime = time.Now()
}

func (x *producerMessage) Delay() time.Duration {
	return x.DeliverAfter
}

func (x *producerMessage) SetDelay(du time.Duration) {
	if du > 0 {
		x.DeliverAfter = du
		x.AddExtra(EXTRA_KIND__DELIVERY_DELAYED.String(), strconv.Itoa(int(du.Milliseconds())))
	}
}

func (x *producerMessage) Underlying() *pulsar.ProducerMessage {
	return &x.ProducerMessage
}

type ConsumerMessage interface {
	mq.HasTopic
	mq.HasPayload
	mq.HasExtra
	mq.HasPartitionKey
	mq.HasOrderingKey
	mq.HasPartitionID
	mq.HasProducer
	mq.HasPublishedAt
	mq.HasConsumedAt
	mq.CanRefreshConsumedAt
	mq.HasLatency
	mq.HasBrokerLatency
	mq.HasRetryCount
	mq.HasUnderlying[pulsar.Message]
}

func NewConsumerMessage(u pulsar.Message) ConsumerMessage {
	m := &consumerMessage{Message: u}
	m.RefreshConsumedAt()
	return m
}

type consumerMessage struct {
	pulsar.Message
	consumedAt time.Time
}

func (x *consumerMessage) Extra() map[string]string {
	return x.Properties()
}

func (x *consumerMessage) ExtraValueOf(k string) (string, bool) {
	if ext := x.Properties(); ext != nil {
		v, ok := ext[k]
		return v, ok
	}
	return "", false
}

func (x *consumerMessage) PartitionKey() string {
	return x.Key()
}

func (x *consumerMessage) PartitionID() int64 {
	return int64(x.ID().PartitionIdx())
}

func (x *consumerMessage) ProducedBy() string {
	return x.ProducerName()
}

func (x *consumerMessage) PublishedAt() time.Time {
	return x.EventTime()
}

func (x *consumerMessage) ConsumedAt() time.Time {
	return x.consumedAt
}

func (x *consumerMessage) RefreshConsumedAt() {
	x.consumedAt = time.Now()
}

func (x *consumerMessage) Latency() time.Duration {
	t1, t2 := x.PublishedAt(), x.ConsumedAt()
	if !t1.IsZero() && !t2.IsZero() && t1.Before(t2) {
		return t2.Sub(t1)
	}
	return 0
}

func (x *consumerMessage) BrokerLatency() time.Duration {
	// x.BrokerPublisherTime() may nil. depends broker feature
	t1, t2 := x.BrokerPublishTime(), x.ConsumedAt()
	if t1 != nil && !t1.IsZero() && !t2.IsZero() && t1.Before(t2) {
		return t2.Sub(*t1)
	}
	return 0
}

func (x *consumerMessage) RetryCount() uint32 {
	return x.RedeliveryCount()
}

func (x *consumerMessage) Underlying() pulsar.Message {
	return x.Message
}
