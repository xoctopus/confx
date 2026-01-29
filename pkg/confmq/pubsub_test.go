package confmq_test

import (
	"context"
	"testing"

	confmq2 "github.com/xoctopus/confx/pkg/confmq"
	. "github.com/xoctopus/x/testx"
)

type MockPubSub struct {
	confmq2.PubSub
	option MockPubOptions
}

type MockPubOptions struct {
	key string
}

func (m *MockPubOptions) OptionScheme() string { return "mock" }

func WithOptionVal(key string) confmq2.OptionApplier {
	return confmq2.OptionApplyFunc(func(o confmq2.Option) {
		if x, ok := o.(*MockPubOptions); ok {
			x.key = key
		}
	})
}

func (ps *MockPubSub) Publish(ctx context.Context, m confmq2.Message, appliers ...confmq2.OptionApplier) error {
	for _, applier := range appliers {
		applier.Apply(&ps.option)
	}
	return nil
}

func TestInjection(t *testing.T) {
	ctx := context.Background()
	_, ok := confmq2.From(ctx)
	Expect(t, ok, BeFalse())

	ps := &MockPubSub{}
	ctx = confmq2.Carry(ps)(ctx)
	Expect(t, confmq2.Must(ctx), Equal[confmq2.PubSub](ps))

	_ = ps.Publish(ctx, nil, WithOptionVal("key"))
	Expect(t, ps.option.key, Equal("key"))
}
