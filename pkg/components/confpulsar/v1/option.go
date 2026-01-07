package confpulsar

import (
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/xoctopus/x/ptrx"

	"github.com/xoctopus/confx/pkg/components/confmq"
	"github.com/xoctopus/confx/pkg/types"
)

// PulsarOption presents pulsar client options
type PulsarOption struct {
	OperationTimeout  types.Duration `url:",default=100ms"`
	ConnTimeout       types.Duration `url:",default=3s"`
	KeepAliveInterval types.Duration `url:",default=1h"`
	MaxConnector      int            `url:",default=10"`
	MaxPending        int            `url:",default=100"`
	MaxBatching       uint           `url:",default=100"`
}

type scheme struct{}

func (*scheme) Scheme() string { return "pulsar" }

func newPubOption(o *PulsarOption, topic string, appliers ...confmq.OptionApplier) *pubOption {
	r := &pubOption{
		options: pulsar.ProducerOptions{
			Topic:               topic,
			SendTimeout:         time.Duration(o.OperationTimeout),
			BatchingMaxMessages: o.MaxBatching,
			MaxPendingMessages:  o.MaxPending,
		},
	}
	for _, applier := range appliers {
		applier.Apply(r)
	}
	r.options.Topic = topic
	return r
}

type pubOption struct {
	scheme
	callback func(confmq.Message, error)
	sync     bool
	options  pulsar.ProducerOptions
}

func WithPublishCallback(callback func(confmq.Message, error)) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if o, ok := opt.(*pubOption); ok {
			o.callback = callback
		}
	})
}

func WithPublishSync() confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if o, ok := opt.(*pubOption); ok {
			o.sync = true
		}
	})
}

func WithPublishTimeout(timeout time.Duration) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if o, ok := opt.(*pubOption); ok {
			o.options.SendTimeout = timeout
		}
	})
}

func WithMaxReconnectToBroker(maxBroker uint) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if o, ok := opt.(*pubOption); ok {
			o.options.MaxReconnectToBroker = ptrx.Ptr(maxBroker)
		}
	})
}

func WithProducerName(name string) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if o, ok := opt.(*pubOption); ok {
			o.options.Name = name
		}
	})
}

func WithPulsarProducerOptions(options pulsar.ProducerOptions) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if o, ok := opt.(*pubOption); ok {
			o.options = options
		}
	})
}

func newSubOption(o *PulsarOption, topic string, appliers ...confmq.OptionApplier) *subOption {
	opt := &subOption{
		scheme:  scheme{},
		options: pulsar.ConsumerOptions{},
	}

	for _, applier := range appliers {
		applier.Apply(opt)
	}
	if len(opt.options.Topics) == 0 && len(opt.options.TopicsPattern) == 0 {
		opt.options.Topic = topic
	}
	if opt.options.SubscriptionName == "" {
		opt.options.SubscriptionName = topic
	}
	return opt
}

type subOption struct {
	scheme
	options pulsar.ConsumerOptions
}

func WithSubTopics(topics ...string) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if x, ok := opt.(*subOption); ok {
			x.options.Topics = topics
		}
	})
}

func WithSubTopicsPattern(pattern string) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if o, ok := opt.(*subOption); ok {
			o.options.TopicsPattern = pattern
		}
	})
}

func WithSubName(name string) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if o, ok := opt.(*subOption); ok {
			o.options.Name = name
		}
	})
}

func WithPulsarConsumerOptions(options pulsar.ConsumerOptions) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if o, ok := opt.(*subOption); ok {
			o.options = options
		}
	})
}
