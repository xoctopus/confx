package confrabbit

import (
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/xoctopus/confx/pkg/types/mq"
)

type ProducerMessage interface {
	mq.HasTopic
	mq.CanSetTopic
	mq.HasPayload
	mq.CanSetPayload
	mq.HasExtra
	mq.CanAppendExtra
	mq.HasPublishedAt
	mq.CanRefreshPublishedAt
	mq.HasDelay
	mq.CanSetDelay
	mq.HasPartitionKey
	mq.CanSetPartitionKey
	mq.HasUnderlying[*amqp.Publishing]
}

func NewProducerMessage(topic string, payload []byte) ProducerMessage {
	m := &producerMessage{
		topic: topic,
		Publishing: amqp.Publishing{
			Body:    payload,
			Headers: make(amqp.Table),
		},
	}
	m.RefreshPublishedAt()
	return m
}

type producerMessage struct {
	topic string
	amqp.Publishing
}

func (x *producerMessage) Topic() string {
	return x.topic
}

func (x *producerMessage) SetTopic(topic string) {
	x.topic = topic
}

func (x *producerMessage) Payload() []byte {
	return x.Body
}

func (x *producerMessage) SetPayload(payload []byte) {
	x.Body = payload
}

func (x *producerMessage) Extra() map[string]string {
	if x.Headers == nil {
		return nil
	}
	ext := make(map[string]string)
	for k, v := range x.Headers {
		if s, ok := v.(string); ok {
			ext[k] = s
		}
	}
	return ext
}

func (x *producerMessage) ExtraValueOf(k string) (string, bool) {
	if x.Headers == nil {
		return "", false
	}
	v, ok := x.Headers[k]
	if ok {
		s, ok := v.(string)
		return s, ok
	}
	return "", false
}

func (x *producerMessage) AddExtra(k, v string) {
	if x.Headers == nil {
		x.Headers = make(amqp.Table)
	}
	x.Headers[k] = v
}

func (x *producerMessage) PartitionKey() string {
	v, _ := x.ExtraValueOf("x-partition-key")
	return v
}

func (x *producerMessage) SetPartitionKey(k string) {
	x.AddExtra("x-partition-key", k)
}

func (x *producerMessage) PublishedAt() time.Time {
	return x.Timestamp
}

func (x *producerMessage) RefreshPublishedAt() {
	x.Timestamp = time.Now()
}

func (x *producerMessage) Delay() time.Duration {
	if x.Headers == nil {
		return 0
	}
	if v, ok := x.Headers["x-delay"]; ok {
		if delayMs, ok := v.(int64); ok {
			return time.Duration(delayMs) * time.Millisecond
		}
		if delayMsStr, ok := v.(string); ok {
			if delayMs, err := strconv.ParseInt(delayMsStr, 10, 64); err == nil {
				return time.Duration(delayMs) * time.Millisecond
			}
		}
	}
	return 0
}

func (x *producerMessage) SetDelay(du time.Duration) {
	if du > 0 {
		x.AddExtra("x-delay", strconv.FormatInt(du.Milliseconds(), 10))
	}
}

func (x *producerMessage) SetDeliveryAt(t time.Time) {
	du := time.Until(t)
	if du > 0 {
		x.SetDelay(du)
	}
}

func (x *producerMessage) Underlying() *amqp.Publishing {
	return &x.Publishing
}

type ConsumerMessage interface {
	mq.HasTopic
	mq.HasPayload
	mq.HasExtra
	mq.HasPartitionKey
	mq.HasPublishedAt
	mq.HasConsumedAt
	mq.CanRefreshConsumedAt
	mq.HasLatency
	mq.HasUnderlying[amqp.Delivery]
}

func NewConsumerMessage(u amqp.Delivery) ConsumerMessage {
	m := &consumerMessage{Delivery: u}
	m.RefreshConsumedAt()
	return m
}

type consumerMessage struct {
	amqp.Delivery
	consumedAt time.Time
}

func (x *consumerMessage) Topic() string {
	return x.RoutingKey
}

func (x *consumerMessage) Payload() []byte {
	return x.Body
}

func (x *consumerMessage) Extra() map[string]string {
	if x.Headers == nil {
		return nil
	}
	ext := make(map[string]string)
	for k, v := range x.Headers {
		if s, ok := v.(string); ok {
			ext[k] = s
		}
	}
	return ext
}

func (x *consumerMessage) ExtraValueOf(k string) (string, bool) {
	if x.Headers == nil {
		return "", false
	}
	v, ok := x.Headers[k]
	if ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}

func (x *consumerMessage) PartitionKey() string {
	v, _ := x.ExtraValueOf("x-partition-key")
	return v
}

func (x *consumerMessage) PublishedAt() time.Time {
	return x.Timestamp
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

func (x *consumerMessage) Underlying() amqp.Delivery {
	return x.Delivery
}
