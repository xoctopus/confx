package confmq_test

import (
	"context"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/components/confmq"
)

type MockPubSub struct {
	confmq.PubSub
	option MockPubOptions
}

type MockPubOptions struct {
	key string
}

func (m *MockPubOptions) Scheme() string { return "mock" }

func WithOptionVal(key string) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(o confmq.Option) {
		if x, ok := o.(*MockPubOptions); ok {
			x.key = key
		}
	})
}

func (ps *MockPubSub) Publish(ctx context.Context, m confmq.Message, appliers ...confmq.OptionApplier) error {
	for _, applier := range appliers {
		applier.Apply(&ps.option)
	}
	return nil
}

func TestInjection(t *testing.T) {
	ctx := context.Background()
	_, ok := confmq.From(ctx)
	Expect(t, ok, BeFalse())

	ps := &MockPubSub{}
	ctx = confmq.Carry(ps)(ctx)
	Expect(t, confmq.Must(ctx), Equal[confmq.PubSub](ps))

	_ = ps.Publish(ctx, nil, WithOptionVal("key"))
	Expect(t, ps.option.key, Equal("key"))
}
