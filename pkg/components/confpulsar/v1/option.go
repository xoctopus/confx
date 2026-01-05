package confpulsar

import (
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/xoctopus/x/ptrx"

	"github.com/xoctopus/confx/pkg/components/confmq"
)

func newDefaultProducerOption(o options, topic string) *producerOption {
	return &producerOption{
		options: pulsar.ProducerOptions{
			Topic:               topic,
			SendTimeout:         time.Duration(o.OperationTimeout),
			BatchingMaxMessages: o.MaxBatching,
			MaxPendingMessages:  o.MaxPending,
		},
	}
}

type producerOption struct {
	callback func(confmq.Message, error)
	sync     bool
	options  pulsar.ProducerOptions
}

func (o producerOption) Role() confmq.Role {
	return confmq.Publisher
}

func (o producerOption) Scheme() string {
	return "pulsar"
}

func WithPublishCallback(callback func(confmq.Message, error)) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if o, ok := opt.(*producerOption); ok {
			o.callback = callback
		}
	})
}

func WithPublishSync() confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if o, ok := opt.(*producerOption); ok {
			o.sync = true
		}
	})
}

func WithPublishTimeout(timeout time.Duration) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if o, ok := opt.(*producerOption); ok {
			o.options.SendTimeout = timeout
		}
	})
}

func WithMaxReconnectToBroker(maxBroker uint) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if o, ok := opt.(*producerOption); ok {
			o.options.MaxReconnectToBroker = ptrx.Ptr(maxBroker)
		}
	})
}
