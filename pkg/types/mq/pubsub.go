package mq

import (
	"context"

	"github.com/xoctopus/x/contextx"
)

// Consumer is the universal subscriber interface for message queues.
// the type parameter M denotes the consumed message type, typically implemented
// by different driver, eg: pulsar, rabbitmq, kafka etc.
type Consumer[M any] interface {
	Acknowledger[M]
	Observer[M]
}

// Observer is a passive consumer interface with no side effects on MQ state.
type Observer[M any] interface {
	// Run starts once consuming loop, processing each message with h.
	// it blocks until ctx canceled or consumer is resource closed and released.
	Run(ctx context.Context, h SubHandler[M]) error
	// Close closes the consumer and releases related resource
	Close() error
}

// Acknowledger provides message acknowledgment and negative acknowledgment.
//
// Used in SubHandler or SubCallback to explicitly control delivery semantics:
// Ack indicates successful processing; Nack indicates failure (typically triggers retry or dead-letter).
type Acknowledger[M any] interface {
	Ack(M) error
	Nack(M) error
}

type AcknowledgerCanDiscard[M any] interface {
	Acknowledger[M]

	// Discard discards message. deliveries message to DLQ directly when biz assures
	// message M cannot be handled, such as message cannot be unmarshalled
	Discard(M) error
}

// Unsubscriber release subscription resource at broker end.
type Unsubscriber interface {
	Unsubscribe() error
}

// Producer is the universal producer interface for message queues.
type Producer[M any] interface {
	// Topic returns the topic name bound to this producer.
	Topic() string
	// Publish publishes payload to the given topic.
	Publish(ctx context.Context, topic string, payload []byte) (M, error)
	// PublishWithKey publishes payload with partition key to the given topic,
	// for partition ordering or load balancing.
	PublishWithKey(ctx context.Context, topic string, key string, payload []byte) (M, error)
	// PublishMessage publishes message
	PublishMessage(context.Context, M) error
	// Close closes the producer and releases connections and related resources.
	Close() error
}

type Factory[M any] = Producer[M]

// PubSub defines the universal MQ client interface using a factory pattern for Producer and Consumer.
// PM is the produced message type, CM is the consumed message type
// instances are created via NewProducer, NewConsumer and configured with OptionApplier
type PubSub[PM any, CM any] interface {
	// NewProducer creates and returns a producer; topic and other options may be set via OptionApplier.
	NewProducer(context.Context, ...OptionApplier) (Producer[PM], error)
	// NewConsumer creates and returns a consumer; topic, subscription name, consume mode, etc. may be set via OptionApplier.
	NewConsumer(context.Context, ...OptionApplier) (Consumer[CM], error)
	// Close closes the pub/sub endpoint and releases underlying connection pools and resources.
	Close() error
}

type Suite[PM any, CM any] interface {
	// Topic returns the topic name bound to this producer.
	Topic() string
	// Publish publishes payload to the given topic.
	Publish(ctx context.Context, topic string, payload []byte) (PM, error)
	// PublishMessage publishes message
	PublishMessage(context.Context, PM) error

	Run(ctx context.Context, h SubHandler[CM]) error
	Acknowledger[CM]

	Close() error
}

// tCtxPubSub serves as the context key type for PubSub, avoiding key collisions across generic instances.
type tCtxPubSub[PM any, CM any] struct{}

// From extracts the PubSub instance from context.
func From[PM any, CM any](ctx context.Context) (PubSub[PM, CM], bool) {
	return contextx.From[tCtxPubSub[PM, CM], PubSub[PM, CM]](ctx)
}

// Must extracts the PubSub instance from context; panics if not present.
func Must[PM any, CM any](ctx context.Context) PubSub[PM, CM] {
	return contextx.Must[tCtxPubSub[PM, CM], PubSub[PM, CM]](ctx)
}

// With injects PubSub into context and returns the new context.
func With[PM any, CM any](ctx context.Context, ps PubSub[PM, CM]) context.Context {
	return contextx.With[tCtxPubSub[PM, CM], PubSub[PM, CM]](ctx, ps)
}

// Carry wraps PubSub as a contextx.Carrier for propagation across process or service boundaries.
func Carry[PM any, CM any](ps PubSub[PM, CM]) contextx.Carrier {
	return contextx.Carry[tCtxPubSub[PM, CM], PubSub[PM, CM]](ps)
}
