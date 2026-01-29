package confmq

import (
	"context"

	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/misc/must"
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

func From(ctx context.Context) (PubSub, bool) {
	ps, ok := ctx.Value(tCtxPubSub{}).(PubSub)
	return ps, ok
}

func Must(ctx context.Context) PubSub {
	return must.BeTrueV(From(ctx))
}

func With(ctx context.Context, ps PubSub) context.Context {
	return context.WithValue(ctx, tCtxPubSub{}, ps)
}

func Carry(ps PubSub) contextx.Carrier {
	return func(ctx context.Context) context.Context {
		return With(ctx, ps)
	}
}
