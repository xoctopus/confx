package mq_test

import (
	"context"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/types/mq"
)

type MockPubSub struct {
	mq.PubSub
	option MockPubOptions
}

type MockPubOptions struct {
	key string
}

func (m *MockPubOptions) OptionScheme() string { return "mock" }

func WithOptionVal(key string) mq.OptionApplier {
	return mq.OptionApplyFunc(func(o mq.Option) {
		if x, ok := o.(*MockPubOptions); ok {
			x.key = key
		}
	})
}

func (ps *MockPubSub) Publish(ctx context.Context, m mq.Message, appliers ...mq.OptionApplier) error {
	for _, applier := range appliers {
		applier.Apply(&ps.option)
	}
	return nil
}

func TestInjection(t *testing.T) {
	ctx := context.Background()
	_, ok := mq.From(ctx)
	Expect(t, ok, BeFalse())

	ps := &MockPubSub{}
	ctx = mq.Carry(ps)(ctx)
	Expect(t, mq.Must(ctx), Equal[mq.PubSub](ps))

	_ = ps.Publish(ctx, nil, WithOptionVal("key"))
	Expect(t, ps.option.key, Equal("key"))
}
