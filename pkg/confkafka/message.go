package confkafka

import (
	"math"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/slicex"

	"github.com/xoctopus/confx/pkg/types/mq"
)

// ProducerMessage composes mq interfaces for Kafka message production.
type ProducerMessage interface {
	mq.HasTopic
	mq.CanSetTopic
	mq.HasPayload
	mq.CanSetPayload
	mq.HasExtra
	mq.CanAppendExtra
	mq.HasExpiredAt
	mq.CanSetExpiredAt
	mq.HasPartitionKey
	mq.CanSetPartitionKey
	mq.HasPublishedAt
	mq.CanRefreshPublishedAt
	mq.HasUnderlying[*kafka.Message]
}

// NewProducerMessage creates a ProducerMessage for the given topic and payload.
func NewProducerMessage(topic string, payload []byte) ProducerMessage {
	m := &producerMessage{Message: kafka.Message{}}
	m.SetTopic(topic)
	m.SetPayload(payload)
	m.RefreshPublishedAt()
	return m
}

type producerMessage struct {
	kafka.Message
}

func (x *producerMessage) Topic() string {
	return x.Message.Topic
}

func (x *producerMessage) SetTopic(topic string) {
	x.Message.Topic = topic
}

func (x *producerMessage) Payload() []byte {
	return x.Message.Value
}

func (x *producerMessage) SetPayload(payload []byte) {
	x.Message.Value = payload
}

func (x *producerMessage) Extra() map[string]string {
	return slicex.Map(x.Message.Headers, func(e kafka.Header) (string, string) {
		return e.Key, string(e.Value)
	})
}

func (x *producerMessage) ExtraValueOf(k string) (string, bool) {
	for _, h := range x.Message.Headers {
		if h.Key == k {
			return string(h.Value), true
		}
	}
	return "", false
}

func (x *producerMessage) AddExtra(k, v string) {
	x.Message.Headers = append(
		x.Message.Headers,
		kafka.Header{Key: k, Value: []byte(v)},
	)
}

func (x *producerMessage) ExpiredAt() int64 {
	expiredAt := int64(math.MaxInt64)
	if val, ok := x.ExtraValueOf(EXTRA_KEY__EXPIRED_AT); ok {
		if v, err := strconv.ParseInt(val, 10, 64); err == nil {
			expiredAt = v
		}
	}
	return expiredAt
}

func (x *producerMessage) SetExpiredAt(expiredAt int64) {
	must.BeTrueF(expiredAt > time.Now().Unix(), "invalid expired timestamp")
	x.AddExtra(EXTRA_KEY__EXPIRED_AT, strconv.FormatInt(expiredAt, 10))
}

func (x *producerMessage) SetExpiredAfter(du time.Duration) {
	x.AddExtra(EXTRA_KEY__EXPIRED_AT, strconv.FormatInt(time.Now().Add(du).Unix(), 10))
}

func (x *producerMessage) PartitionKey() string {
	return string(x.Message.Key)
}

func (x *producerMessage) SetPartitionKey(k string) {
	x.Message.Key = []byte(k)
}

func (x *producerMessage) PublishedAt() time.Time {
	return x.Message.Time
}

func (x *producerMessage) RefreshPublishedAt() {
	x.Message.Time = time.Now()
}

func (x *producerMessage) Underlying() *kafka.Message {
	return &x.Message
}

// ConsumerMessage composes mq interfaces for Kafka message consumption.
type ConsumerMessage interface {
	mq.HasTopic
	mq.HasPayload
	mq.HasExtra
	mq.HasExpiredAt
	mq.HasPartitionKey
	mq.HasOrderingKey
	mq.HasPartitionID
	mq.HasOffset
	mq.HasProducer
	mq.HasPublishedAt
	mq.HasConsumedAt
	mq.CanRefreshConsumedAt
	mq.HasLatency
	mq.HasRetryCount
	mq.HasUnderlying[kafka.Message]
}

// NewConsumerMessage wraps a kafka.Message as ConsumerMessage.
func NewConsumerMessage(u kafka.Message) ConsumerMessage {
	m := &consumerMessage{Message: u}
	m.RefreshConsumedAt()
	return m
}

type consumerMessage struct {
	kafka.Message
	consumedAt time.Time
}

func (x *consumerMessage) Topic() string {
	return x.Message.Topic
}

func (x *consumerMessage) Payload() []byte {
	return x.Message.Value
}

func (x *consumerMessage) Extra() map[string]string {
	return slicex.Map(x.Message.Headers, func(e kafka.Header) (string, string) {
		return e.Key, string(e.Value)
	})
}

func (x *consumerMessage) ExtraValueOf(k string) (string, bool) {
	for _, h := range x.Message.Headers {
		if h.Key == k {
			return string(h.Value), true
		}
	}
	return "", false
}

func (x *consumerMessage) ExpiredAt() int64 {
	expiredAt := int64(math.MaxInt64)
	if val, ok := x.ExtraValueOf(EXTRA_KEY__EXPIRED_AT); ok {
		if v, err := strconv.ParseInt(val, 10, 64); err == nil {
			expiredAt = v
		}
	}
	return expiredAt
}

func (x *consumerMessage) PartitionKey() string {
	return string(x.Message.Key)
}

func (x *consumerMessage) OrderingKey() string {
	return string(x.Message.Key)
}

func (x *consumerMessage) PartitionID() int64 {
	return int64(x.Message.Partition)
}

func (x *consumerMessage) Offset() int64 {
	return x.Message.Offset
}

// ProducedBy has no Kafka equivalent; returns empty string.
func (x *consumerMessage) ProducedBy() string {
	return ""
}

func (x *consumerMessage) PublishedAt() time.Time {
	return x.Message.Time
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

func (x *consumerMessage) RetryCount() uint32 {
	if val, ok := x.ExtraValueOf(EXTRA_KEY__RETRY_COUNT); ok {
		if v, err := strconv.ParseUint(val, 10, 32); err == nil {
			return uint32(v)
		}
	}
	return 0
}

func (x *consumerMessage) Underlying() kafka.Message {
	return x.Message
}
