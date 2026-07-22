package confrabbit

import (
	"time"

	"github.com/wagslane/go-rabbitmq"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/slicex"
	"github.com/xoctopus/x/textx"

	"github.com/xoctopus/confx/pkg/conftls"
	"github.com/xoctopus/confx/pkg/types"
	"github.com/xoctopus/confx/pkg/types/mq"
)

type Option struct {
	Addresses []string
	Shuffle   bool

	ReconnectInterval types.Duration `url:",default=5s"`
	Vhost             string         `url:",default=/"`
	TLS               conftls.X509KeyPair

	PubTimeout types.Duration `url:",default=2s"`

	defaultPubOption *PubOption
	defaultSubOption *SubOption
}

func (o *Option) SetDefault() {
	_ = must.NoErrorV(textx.SetDefault(o))

	if o.defaultPubOption == nil {
		o.defaultPubOption = &PubOption{
			timeout: time.Duration(o.PubTimeout),
		}
	}
	if o.defaultSubOption == nil {
		o.defaultSubOption = &SubOption{
			worker:     16,
			bufferSize: 1024,
			mode:       mq.Concurrent,
			hasher:     mq.Fnv,
		}
	}
}

func (o *Option) URLs(main string) []string {
	return slicex.Unique(append([]string{main}, o.Addresses...))
}

func (o *Option) ClientOptions() []func(*rabbitmq.ConnectionOptions) {
	var opts []func(*rabbitmq.ConnectionOptions)

	cfg := rabbitmq.Config{
		Vhost: o.Vhost,
	}

	if !o.TLS.IsZero() {
		must.NoError(o.TLS.Init())
		cfg.TLSClientConfig = o.TLS.Config()
	}

	opts = append(opts, rabbitmq.WithConnectionOptionsConfig(cfg))

	return opts
}

func (o *Option) PubOption(appliers ...mq.OptionApplier) *PubOption {
	opt := *o.defaultPubOption
	for _, applier := range appliers {
		applier.Apply(&opt)
	}
	return &opt
}

func (o *Option) SubOption(appliers ...mq.OptionApplier) *SubOption {
	opt := *o.defaultSubOption
	for _, applier := range appliers {
		applier.Apply(&opt)
	}
	return &opt
}

// WithPubTopic sets the topic for publisher.
// NOTE: In RabbitMQ, Topic concept is mapped to 'Routing Key'.
func WithPubTopic(topic string) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.topic = topic
		}
	})
}

// WithPubRoutingKey is an alias of WithPubTopic for clearer intention in RabbitMQ.
func WithPubRoutingKey(routingKey string) mq.OptionApplier {
	return WithPubTopic(routingKey)
}

// WithPubExchange explicit declares the destination Exchange for the publisher.
// 'name' is the exchange name, 'kind' could be "direct", "fanout", "topic" or "headers".
func WithPubExchange(name, kind string) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.exchangeName, x.exchangeKind = name, kind
			x.options = append(x.options,
				rabbitmq.WithPublisherOptionsExchangeName(name),
				rabbitmq.WithPublisherOptionsExchangeKind(kind),
			)
		}
	})
}

func WithSyncPublish() mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.sync = true
		}
	})
}

func WithPublishCallback(f mq.AsyncPubCallback[ProducerMessage]) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.callback = f
		}
	})
}

func WithPubTimeout(d time.Duration) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.timeout = d
		}
	})
}

func WithRabbitPublisherOptions(opts ...func(*rabbitmq.PublisherOptions)) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.options = append(x.options, opts...)
		}
	})
}

type PubOption struct {
	topic    string
	sync     bool
	callback mq.AsyncPubCallback[ProducerMessage]

	exchangeName string
	exchangeKind string

	timeout time.Duration

	options []func(*rabbitmq.PublisherOptions)
}

func (*PubOption) OptionScheme() string { return "rabbitmq" }

func WithSubQueue(queue string) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.queue = queue
		}
	})
}

// WithSubExchange explicit declares the source Exchange for the consumer,
// and implicitly binds the Queue to this Exchange using the Queue name as RoutingKey.
func WithSubExchange(name, kind string) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.options = append(x.options,
				rabbitmq.WithConsumerOptionsExchangeName(name),
				rabbitmq.WithConsumerOptionsExchangeKind(kind),
			)
		}
	})
}

// WithSubRoutingKey explicit sets the routing key for consumer binding.
// Note: If you want to use wildcards like `*.info` or `#`, use this option
// along with `WithSubExchange(name, "topic")`.
func WithSubRoutingKey(routingKey string) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.options = append(x.options,
				rabbitmq.WithConsumerOptionsRoutingKey(routingKey),
			)
		}
	})
}

func WithSubConsumeMode(mode mq.ConsumeHandleMode) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.mode = mode
		}
	})
}

func WithSubWorker(n uint16) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.worker = n
		}
	})
}

func WithSubBufferSize(n uint16) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.bufferSize = n
		}
	})
}

func WithSubHasher(h mq.Hasher) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.hasher = h
		}
	})
}

func WithSubDisableAutoAck() mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.disableAutoAck = true
		}
	})
}

func WithSubCallback(f mq.SubCallback[ConsumerMessage]) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.callback = f
		}
	})
}

func WithRabbitConsumerOptions(opts ...func(*rabbitmq.ConsumerOptions)) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.options = append(x.options, opts...)
		}
	})
}

type SubOption struct {
	queue          string
	mode           mq.ConsumeHandleMode
	worker         uint16
	hasher         mq.Hasher
	bufferSize     uint16
	callback       mq.SubCallback[ConsumerMessage]
	disableAutoAck bool
	options        []func(options *rabbitmq.ConsumerOptions)
}

func (*SubOption) OptionScheme() string { return "rabbitmq" }
