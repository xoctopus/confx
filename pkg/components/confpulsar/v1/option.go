package confpulsar

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/apache/pulsar-client-go/pulsar/backoff"
	"github.com/apache/pulsar-client-go/pulsar/log"

	"github.com/xoctopus/confx/pkg/components/confmq"
	"github.com/xoctopus/confx/pkg/types"
)

// PulsarOption presents pulsar client options
type PulsarOption struct {
	// ConnectionTimeout [Client] establishment timeout
	ConnectionTimeout types.Duration `url:",default=1s"`
	// ConnectionMaxIdleTime [Client] release the connection if it is not
	// used for more than ConnectionMaxIdleTime. default is 30 minutes
	ConnectionMaxIdleTime types.Duration `url:",default=30m"`
	// OperationTimeout [Client] producer-create, subscribe and unsubscribe
	// operations will be retried until this interval
	OperationTimeout types.Duration `url:",default=3s"`
	// KeepAliveInterval [Client] the ping send and check interval
	KeepAliveInterval types.Duration `url:",default=1m"`
	// MaxConnectionsPerBroker [Client] max connections to a single broker
	MaxConnectionsPerBroker int `url:",default=10"`

	// SendTimeout [PUB] specifies the timeout for a message from sent to
	// acknowledged by the server
	SendTimeout types.Duration `url:",default=2s"`
	// DisableBlockIfQueueFull [PUB] controls whether Send and SendAsync block
	// when producer's message queue is full.
	DisableBlockIfQueueFull bool `url:",default=false"`
	// MaxPendingMessages [PUB] specifies the max size of the queue holding
	MaxPendingMessages int `url:",default=500"`
	// DisableCompress [PUB] specifies if disable message compression, if it is
	// enabled use LZ4 compress type
	DisableCompress bool `url:",default=false"`
	// BatchingMaxMessages [PUB] specifies the max messages permitted in a batch
	BatchingMaxMessages uint
	// EnablePubShared [PUB] if disabled, publisher is required exclusive access
	// for producer. failed immediately if there's already a producer connected.
	EnablePubShared bool `url:",default=false"`

	// EnableSubShared [SUB] if disabled, there can be only 1 consumer on the same
	// topic with the same subscription name
	EnableSubShared bool `url:",default=true"`
	// EnableRetryNack [SUB] if enabled, NACKed message will be redelivered after
	// NackRetryInterval max MaxNackRetry times. if reached MaxNackRetry times,
	// the message filled to global DLQ
	EnableRetryNack bool `url:",default=true"`
	// NackRetryInterval [SUB] retry nack message interval
	NackRetryInterval types.Duration `url:",default=1m"`
	// MaxNackRetry [SUB] max retry times for nack message
	MaxNackRetry uint32 `url:",default=3"`

	// internal
	defaultPubOption *PubOption
	defaultSubOption *SubOption
}

func (o *PulsarOption) SetDefault() {
	if o.defaultPubOption == nil {
		o.defaultPubOption = &PubOption{}
	}
	if !o.defaultPubOption._initialized {
		compressMode, compressLevel := pulsar.NoCompression, pulsar.Default
		if !o.DisableCompress {
			compressMode = pulsar.LZ4
			compressLevel = pulsar.Default
		}
		disableBatching := false
		if o.BatchingMaxMessages == 0 {
			disableBatching = true
		}
		accessMode := pulsar.ProducerAccessModeExclusive
		if o.EnablePubShared {
			accessMode = pulsar.ProducerAccessModeShared
		}

		o.defaultPubOption = &PubOption{
			options: pulsar.ProducerOptions{
				SendTimeout:             time.Duration(o.OperationTimeout),
				DisableBlockIfQueueFull: o.DisableBlockIfQueueFull,
				MaxPendingMessages:      o.MaxPendingMessages,
				CompressionType:         compressMode,
				CompressionLevel:        compressLevel,
				DisableBatching:         disableBatching,
				BatchingMaxMessages:     o.BatchingMaxMessages,
				BackOffPolicyFunc:       func() backoff.Policy { return &backoff.DefaultBackoff{} },
				ProducerAccessMode:      accessMode,
			},
		}
		o.defaultPubOption._initialized = true
	}
	if o.defaultSubOption == nil {
		o.defaultSubOption = &SubOption{}
	}
	if !o.defaultSubOption._initialized {
		subMode := pulsar.Exclusive
		if o.EnableSubShared {
			subMode = pulsar.Shared
		}

		dlq := (*pulsar.DLQPolicy)(nil)
		if o.EnableRetryNack {
			dlq = &pulsar.DLQPolicy{MaxDeliveries: o.MaxNackRetry}
		}

		o.defaultSubOption = &SubOption{
			options: pulsar.ConsumerOptions{
				Type:                           subMode,
				EventListener:                  nil,
				DLQ:                            dlq,
				RetryEnable:                    o.EnableRetryNack,
				NackRedeliveryDelay:            time.Duration(o.NackRetryInterval),
				EnableDefaultNackBackoffPolicy: true,
			},
		}
		o.defaultSubOption._initialized = true
	}
}

func (o *PulsarOption) ClientOption(url string) pulsar.ClientOptions {
	l := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelWarn}))

	return pulsar.ClientOptions{
		URL:                     url,
		ConnectionTimeout:       time.Duration(o.ConnectionTimeout),
		ConnectionMaxIdleTime:   time.Duration(o.ConnectionMaxIdleTime),
		OperationTimeout:        time.Duration(o.OperationTimeout),
		KeepAliveInterval:       time.Duration(o.KeepAliveInterval),
		MaxConnectionsPerBroker: o.MaxConnectionsPerBroker,
		Logger:                  log.NewLoggerWithSlog(l),
	}
}

func (o *PulsarOption) PubOption(topic string, appliers ...confmq.OptionApplier) *PubOption {
	opt := *o.defaultPubOption
	for _, applier := range appliers {
		applier.Apply(&opt)
	}
	opt.options.Topic = topic
	return &opt
}

func (o *PulsarOption) SubOption(topic string, appliers ...confmq.OptionApplier) *SubOption {
	opt := *o.defaultSubOption
	for _, applier := range appliers {
		applier.Apply(&opt)
	}
	opt.options.Topic = topic
	if opt.options.SubscriptionName == "" {
		opt.options.SubscriptionName = topic
	}
	return &opt
}

type PubOption struct {
	_initialized bool
	// failover when sync send mode disabled. failover will be called when message
	// sent succeed or failed
	failover func(confmq.Message, error)
	// sync decides use Send or SendAsync in pulsar client
	sync bool
	// options pulsar producer option
	options pulsar.ProducerOptions
}

func (*PubOption) OptionScheme() string { return "pulsar" }

func WithPublishFailover(f func(confmq.Message, error)) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.failover = f
		}
	})
}

func WithSyncPublish() confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.sync = true
		}
	})
}

func WithPubSendTimeout(d time.Duration) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.options.SendTimeout = d
		}
	})
}

func WithPubEnableBlockIfQueueFull() confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.options.DisableBlockIfQueueFull = true
		}
	})
}

func WithPubMaxPendingMessages(n int) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.options.MaxPendingMessages = n
		}
	})
}

func WithPubEnableCompression() confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.options.CompressionType = pulsar.LZ4
			x.options.CompressionLevel = pulsar.Default
		}
	})
}

func WithPubBatchingMaxMessages(n uint) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.options.DisableBatching = false
			x.options.BatchingMaxMessages = n
		}
	})
}

func WithPubAccessMode(m pulsar.ProducerAccessMode) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.options.ProducerAccessMode = pulsar.ProducerAccessModeShared
		}
	})
}

func WithPublisherOptions(o pulsar.ProducerOptions) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.options = o
		}
	})
}

type SubOption struct {
	_initialized bool
	// failover it is called when invalid message received or custom handler
	// panicked. if needed consumer set this attribute to hook failed case.
	failover func(context.Context, error)
	// options pulsar consumer options
	options pulsar.ConsumerOptions
}

func (*SubOption) OptionScheme() string { return "pulsar" }

func WithSubFailover(f func(context.Context, error)) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.failover = f
		}
	})
}

func WithSubscriptionName(name string) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.options.SubscriptionName = name
		}
	})
}

func WithSubscriptionType(t pulsar.SubscriptionType) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.options.Type = t
		}
	})
}

func WithSubEnableRetryNack(retryDelay time.Duration, maxRetry uint32) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.options.RetryEnable = true
			x.options.NackRedeliveryDelay = retryDelay
			if x.options.DLQ == nil {
				x.options.DLQ = &pulsar.DLQPolicy{}
			}
			x.options.DLQ.MaxDeliveries = maxRetry
		}
	})
}

func WithPulsarConsumerOptions(options pulsar.ConsumerOptions) confmq.OptionApplier {
	return confmq.OptionApplyFunc(func(opt confmq.Option) {
		if o, ok := opt.(*SubOption); ok {
			o.options = options
		}
	})
}
