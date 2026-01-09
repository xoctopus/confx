package confmq

import (
	"context"

	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/misc/must"
)

type Subscriber interface {
	Topic() string

	Run(context.Context, func(context.Context, Message) error) <-chan error
	Close() error
}

type PubSub interface {
	Publish(ctx context.Context, msg Message, options ...OptionApplier) error
	Subscribe(ctx context.Context, topic string, options ...OptionApplier) (Subscriber, error)
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
