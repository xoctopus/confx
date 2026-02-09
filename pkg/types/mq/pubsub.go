package mq

import (
	"context"

	"github.com/xoctopus/x/contextx"
)

type Subscriber interface {
	// Run starts consuming and handling messages with h
	Run(ctx context.Context, h func(context.Context, Message) error) <-chan error
	// Close closes this subscriber
	Close()
}

type Publisher interface {
	Topic() string
	Publish(context.Context, any) error
	PublishMessage(context.Context, Message) error
	Close()
}

type PubSub interface {
	// Publisher returns a publisher
	Publisher(ctx context.Context, options ...OptionApplier) (Publisher, error)
	// Subscriber returns a subscriber
	Subscriber(ctx context.Context, options ...OptionApplier) (Subscriber, error)
	// Close closes pub/sub endpoint
	Close() error
}

type tCtxPubSub struct{}

var (
	From  = contextx.From[tCtxPubSub, PubSub]
	Must  = contextx.Must[tCtxPubSub, PubSub]
	With  = contextx.With[tCtxPubSub, PubSub]
	Carry = contextx.Carry[tCtxPubSub, PubSub]
)
