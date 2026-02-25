package mq

import "time"

type HasTopic interface {
	Topic() string
}

type CanSetTopic interface {
	SetTopic(string)
}

type HasPayload interface {
	Payload() []byte
}

type CanSetPayload interface {
	SetPayload([]byte)
}

type HasSequenceID interface {
	SequenceID() int64
}

type CanSetSequenceID interface {
	SetSequenceID(int64)
}

type HasExtra interface {
	// Extra is used to extend message info, eg: TraceID, RetryCount etc.
	Extra() map[string]string
	ExtraValueOf(string) (string, bool)
}

type CanAppendExtra interface {
	AddExtra(k string, v string)
}

type HasTags interface {
	Tags() []string
}

type CanAppendTags interface {
	AddTags(...string)
}

type HasExpiredAt interface {
	ExpiredAt() int64
}

type CanSetExpiredAt interface {
	// SetExpiredAt use epoch second timestamp
	SetExpiredAt(int64)
	// SetExpiredAfter base now
	SetExpiredAfter(time.Duration)
}

// HasPartitionKey is used during message production to retrieve or specify
// the biz partition key for sharding.
type HasPartitionKey interface {
	// PartitionKey biz key used for partition hashing
	PartitionKey() string
}

type CanSetPartitionKey interface {
	SetPartitionKey(string)
}

// HasOrderingKey is used during message consumption to identify the biz
// key that governs processing order
type HasOrderingKey interface {
	OrderingKey() string
}

type CanSetOrderingKey interface {
	SetOrderingKey(string)
}

type HasDelay interface {
	Delay() time.Duration
}

type CanSetDelay interface {
	SetDelay(time.Duration)
	SetDeliveryAt(time.Time)
}

type HasRetryCount interface {
	RetryCount() uint32
}

type CanSetRetryCount interface {
	AddRetryCount()
	SetRetryCount(uint32)
}

// HasBacklog presents backlog messages when consuming
type HasBacklog interface {
	Backlog() int64
}

type CanSetBacklog interface {
	SetBacklog(int64)
}

type HasPartitionID interface {
	PartitionID() int64
}

type CanSetPartitionID interface {
	SetPartitionID(uint64)
}

type HasOffset interface {
	Offset() int64
}

type CanSetOffset interface {
	SetOffset(int64)
}

type HasProducer interface {
	ProducedBy() string
}

type CanSetProducer interface {
	SetProducer(string)
}

type HasPublishedAt interface {
	PublishedAt() time.Time
}

type CanSetPublishedAt interface {
	SetPublishedAt(time.Time)
}

type CanRefreshPublishedAt interface {
	RefreshPublishedAt()
}

type HasConsumedAt interface {
	ConsumedAt() time.Time
}

type CanSetConsumedAt interface {
	SetConsumedAt(time.Time)
}

type CanRefreshConsumedAt interface {
	RefreshConsumedAt()
}

type HasLatency interface {
	// Latency duration from event time(first published) to logic subscribed
	Latency() time.Duration
}

type HasBrokerLatency interface {
	// BrokerLatency duration from broker published to logic subscribed
	BrokerLatency() time.Duration
}

// HasUnderlying returns or retrieves the raw underlying message value.
// this is typically used to access driver-specific message type provided by the
// MQ implementation (e.g. pulsar.Message, kafka.Message, etc.).
type HasUnderlying[T any] interface {
	Underlying() T
}

type CanSetUnderlying[T any] interface {
	SetUnderlying(T)
}
