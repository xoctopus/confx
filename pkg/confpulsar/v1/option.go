package confpulsar

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/apache/pulsar-client-go/pulsar/backoff"
	"github.com/apache/pulsar-client-go/pulsar/log"

	"github.com/xoctopus/confx/pkg/types"
	"github.com/xoctopus/confx/pkg/types/mq"
)

// Option presents pulsar client options and default pub/sub options. it can be
// overridden by option applier when call Endpoint.Publish and Endpoint.Subscribe
type Option struct {
	// Tenant represents the top-level Pulsar tenant.
	// It is the highest isolation boundary in Pulsar, usually used to separate
	// environments (prod/test/dev) or organizations.
	// It maps to the `<tenant>` part of `persistent://<tenant>/<namespace>/<topic>`.
	Tenant string `url:",default=public"`

	// Namespace represents the logical domain under a tenant.
	// It is the main unit for policy and resource management in Pulsar, such as:
	// retention, TTL, backlog quota, permissions, and schema enforcement.
	// It maps to the `<namespace>` part of `persistent://<tenant>/<namespace>/<topic>`.
	Namespace string `url:",default=default"`

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
	// WorkerSize defines the concurrency level for message consumption.
	// Behavior based on ConsumeMode:
	// eg:
	//	- mq.GlobalOrdered: Forced to 1 to ensure strict sequential processing.
	//	- mq.PartitionOrdered: Messages are dispatched to specific workers based on
	//	  a hash of the partition key, ensuring order within the same key.
	//	- mq.Concurrent: messages are distributed across all workers (e.g., round-robin)
	//	  to maximize throughput.
	WorkerSize uint16 `url:",default=16"`
	// WorkerBufferSize [SUB] will prefetch message from broker for improves
	// consumption throughput and reducing wait
	WorkerBufferSize uint16 `url:",default=64"`

	// DisablePersistent [TOPIC] if disable persistent. set true the topic prefix
	// use `non-persistent`.
	// eg:
	// persistent://tenant/namespace/topic
	// non-persistent://tenant/namespace/topic
	DisablePersistent bool `url:",default=false"`

	// prefix persistent url prefix
	prefix string
	// defaultPubOption default publisher option
	defaultPubOption *PubOption
	// defaultPubOption default subscriber option
	defaultSubOption *SubOption
}

func (o *Option) SetDefault() {
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
				BackOffPolicyFunc:       func() backoff.Policy { return backoff.NewDefaultBackoff() },
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
			worker:     o.WorkerSize,
			hasher:     mq.CRC,
			bufferSize: o.WorkerBufferSize,
		}
		o.defaultSubOption._initialized = true
	}

	o.prefix = "persistent"
	if o.DisablePersistent {
		o.prefix = "non-persistent"
	}
}

func (o *Option) String() string {
	return fmt.Sprintf("%s://%s/%s/", o.prefix, o.Tenant, o.Namespace)
}

func (o *Option) Topic(topic string) string {
	return fmt.Sprintf("%s://%s/%s/%s", o.prefix, o.Tenant, o.Namespace, topic)
}

func (o *Option) ClientOption(url string) pulsar.ClientOptions {
	l := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

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

	if opt.mode == mq.GlobalOrdered {
		opt.worker = 1
	}
	if opt.worker == 0 {
		opt.worker = 16
	}
	if opt.bufferSize == 0 {
		opt.bufferSize = 256
	}
	if opt.mode == mq.PartitionOrdered && opt.hasher == nil {
		opt.hasher = mq.CRC
	}

	return &opt
}

type PubOption struct {
	_initialized bool
	// callback when async send mode enabled. callback will be called when message
	// sent completed
	callback func(mq.Message, error)
	// sync decides use Send or SendAsync in pulsar client
	sync bool
	// options pulsar producer option
	options pulsar.ProducerOptions
}

func (o *PubOption) Options() pulsar.ProducerOptions {
	return o.options
}

func (*PubOption) OptionScheme() string { return "pulsar" }

func WithPublishCallback(f func(mq.Message, error)) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.callback = f
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

func WithPubTopic(topic string) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.options.Topic = topic
		}
	})
}

func WithPubSendTimeout(d time.Duration) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.options.SendTimeout = d
		}
	})
}

func WithPubEnableBlockIfQueueFull() mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.options.DisableBlockIfQueueFull = true
		}
	})
}

func WithPubMaxPendingMessages(n int) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.options.MaxPendingMessages = n
		}
	})
}

func WithPubEnableCompression() mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.options.CompressionType = pulsar.LZ4
			x.options.CompressionLevel = pulsar.Default
		}
	})
}

func WithPubBatchingMaxMessages(n uint) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.options.DisableBatching = false
			x.options.BatchingMaxMessages = n
		}
	})
}

func WithPubAccessMode(m pulsar.ProducerAccessMode) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.options.ProducerAccessMode = m
		}
	})
}

func WithPublisherOptions(o pulsar.ProducerOptions) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*PubOption); ok {
			x.options = o
		}
	})
}

type SubOption struct {
	_initialized bool
	// disableAutoAck disable auto ack. if this option is set true, message ack
	// should be handled by callback.
	disableAutoAck bool
	// callback it is called when message handled
	callback func(pulsar.Consumer, pulsar.Message, mq.Message, error)
	// handler process mq.Message
	handler mq.Handler
	// worker specifies the consumer concurrency level. the default is defined
	// in Option.WorkerSize (16 by default).
	worker uint16
	// bufferSize
	bufferSize uint16
	// hasher helps to hash message partition key
	hasher mq.Hasher
	// mode consumer handling mode
	mode mq.ConsumeMode
	// options pulsar consumer options
	options pulsar.ConsumerOptions
}

func (*SubOption) OptionScheme() string { return "pulsar" }

func (o *SubOption) Options() pulsar.ConsumerOptions {
	return o.options
}

// WithSubDisableAutoAck enables auto ack. when message received from mq, ack will
// be performed immediately.
func WithSubDisableAutoAck() mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.disableAutoAck = true
		}
	})
}

// WithSubCallback set subscriber's callback when message is handled.
func WithSubCallback(f func(pulsar.Consumer, pulsar.Message, mq.Message, error)) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.callback = f
		}
	})
}

func WithSubTopic(topics ...string) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			if len(topics) > 0 {
				if x.options.SubscriptionName == "" {
					x.options.SubscriptionName = topics[0]
				}
				if len(topics) == 1 {
					x.options.Topic = topics[0]
				}
				if len(topics) > 1 {
					x.options.Topics = topics
				}
			}
		}
	})
}

func WithSubTopicPattern(pattern string) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.options.TopicsPattern = pattern
		}
	})
}

func WithSubName(name string) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.options.SubscriptionName = name
		}
	})
}

func WithSubType(t pulsar.SubscriptionType) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.options.Type = t
		}
	})
}

func WithSubEnableRetryNack(retryDelay time.Duration, maxRetry uint32) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.options.RetryEnable = true
			if maxRetry > 1 {
				x.options.NackBackoffPolicy = &nackBackoffPolicy{retryDelay, maxRetry}
			} else {
				x.options.NackRedeliveryDelay = retryDelay
			}
			if x.options.DLQ == nil {
				x.options.DLQ = &pulsar.DLQPolicy{}
			}
			x.options.DLQ.MaxDeliveries = maxRetry
			x.options.DLQ.DeadLetterTopic = x.options.Topic + "_DLQ"
		}
	})
}

func WithSubWorkerSize(n uint16) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.bufferSize = n
		}
	})
}

func WithSubWorkerBufferSize(n uint16) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.worker = n
		}
	})
}

func WithSubOrderedKeyHasher(h mq.Hasher) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.hasher = h
		}
	})
}

func WithSubConsumingMode(mode mq.ConsumeMode) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if x, ok := opt.(*SubOption); ok {
			x.mode = mode
		}
	})
}

func WithPulsarConsumerOptions(options pulsar.ConsumerOptions) mq.OptionApplier {
	return mq.OptionApplyFunc(func(opt mq.Option) {
		if o, ok := opt.(*SubOption); ok {
			o.options = options
		}
	})
}

type nackBackoffPolicy struct {
	retryDelay time.Duration
	maxRetry   uint32
}

func (p *nackBackoffPolicy) Next(count uint32) time.Duration {
	count = max(1, count)
	count = min(count, p.maxRetry)
	return min(p.retryDelay*time.Duration(count), time.Minute*10)
}
