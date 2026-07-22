package confrabbit

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/wagslane/go-rabbitmq"
	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/confx/pkg/types"
	"github.com/xoctopus/confx/pkg/types/liveness"
	"github.com/xoctopus/confx/pkg/types/mq"
)

type Endpoint struct {
	types.Endpoint[Option]

	client *rabbitmq.Conn
	closed atomic.Bool

	mq.ResourceManager `env:"-"`
}

var _ mq.PubSub[ProducerMessage, ConsumerMessage] = (*Endpoint)(nil)

func (e *Endpoint) SetDefault() {
	if e.Endpoint.Address == "" {
		e.Endpoint.Address = "amqp://localhost"
	}
	if e.Option.Vhost == "" {
		e.Option.Vhost = "/"
	}
	if e.ResourceManager == nil {
		e.ResourceManager = mq.NewResourceManager()
	}
	e.Option.SetDefault()
	e.closed.Store(false)
}

func (e *Endpoint) Init(ctx context.Context) error {
	if err := e.Endpoint.Init(); err != nil {
		return err
	}

	if e.client == nil {
		var (
			client   *rabbitmq.Conn
			err      error
			appliers = e.Option.ClientOptions()
			urls     = e.Option.URLs(e.Address)
		)

		if len(urls) > 1 {
			resolver := rabbitmq.NewStaticResolver(urls, e.Option.Shuffle)
			client, err = rabbitmq.NewClusterConn(resolver, appliers...)
		} else {
			client, err = rabbitmq.NewConn(e.Address, appliers...)
		}

		if err != nil {
			return err
		}
		e.client = client
	}

	return e.LivenessCheck(ctx).FailureReason()
}

func (e *Endpoint) LivenessCheck(ctx context.Context) (v liveness.Result) {
	v = liveness.NewLivenessData()
	v.Start()

	if e.closed.Load() || e.client == nil {
		v.End(codex.New(ECODE__CLI_CLOSED))
		return
	}

	var (
		err  error
		p    mq.Producer[ProducerMessage]
		c    mq.Consumer[ConsumerMessage]
		body = []byte(uuid.NewString())
	)
	defer func() {
		v.End(err)
	}()

	// Readiness check via a temporary exclusive queue & exchange
	exchange := "liveness_exchange"
	topic := "liveness"
	queue := uuid.NewString()

	// Create temporary publisher
	p, err = e.NewProducer(ctx,
		WithPubExchange(exchange, "direct"),
		WithSyncPublish(),
		WithPubRoutingKey(topic),
		WithRabbitPublisherOptions(
			rabbitmq.WithPublisherOptionsExchangeAutoDelete,
			rabbitmq.WithPublisherOptionsExchangeDeclare,
		),
	)
	if err != nil {
		return
	}
	defer func() { _ = p.Close() }()

	// Create temporary consumer
	c, err = e.NewConsumer(ctx,
		WithSubQueue(queue),
		WithSubExchange(exchange, "direct"),
		WithSubWorker(1),
		WithSubRoutingKey(topic),
		WithRabbitConsumerOptions(
			rabbitmq.WithConsumerOptionsExchangeAutoDelete,
			rabbitmq.WithConsumerOptionsExchangeDeclare,
			rabbitmq.WithConsumerOptionsQueueAutoDelete,
			rabbitmq.WithConsumerOptionsQueueExclusive,
		),
	)
	if err != nil {
		return
	}
	defer func() { _ = c.Close() }()

	sig := make(chan struct{}, 1)
	go func() {
		_ = c.Run(ctx, func(_ context.Context, msg ConsumerMessage) error {
			if string(msg.Payload()) == string(body) {
				select {
				case sig <- struct{}{}:
				default:
				}
			}
			return nil
		})
	}()

	// Wait for echo
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for {
		// Publish test message
		msg := NewProducerMessage(topic, body)
		err = p.PublishMessage(ctx, msg)
		if err != nil {
			return
		}

		select {
		case <-sig:
			return
		case <-ctx.Done():
			err = fmt.Errorf("echo timeout in 5 seconds")
			return
		case <-time.After(500 * time.Millisecond):
			// Retry publish
		}
	}
}

func (e *Endpoint) NewProducer(ctx context.Context, options ...mq.OptionApplier) (_ mq.Producer[ProducerMessage], err error) {
	_, log := logx.Enter(ctx)
	var (
		p   *rabbitmq.Publisher
		x   *producer
		opt = e.Option.PubOption(options...)
	)
	defer func() {
		if err != nil {
			log.Error(err)
		} else {
			e.AddProducer(x)
			log.Info("pub created")
		}
		log.End()
	}()

	if e.closed.Load() || e.client == nil {
		return nil, codex.New(ECODE__CLI_CLOSED)
	}

	log = log.With("topic", opt.topic, "exchange", opt.exchangeName, "kind", opt.exchangeKind)
	p, err = rabbitmq.NewPublisher(e.client, opt.options...)
	if err != nil {
		return
	}

	x = &producer{
		cli:      e,
		pub:      p,
		log:      logx.NewStd().With("topic", opt.topic),
		topic:    opt.topic,
		exchange: opt.exchangeName,
		timeout:  opt.timeout,
		sync:     opt.sync,
		callback: opt.callback,
	}
	return x, nil
}

func (e *Endpoint) NewConsumer(ctx context.Context, options ...mq.OptionApplier) (_ mq.Consumer[ConsumerMessage], err error) {
	_, log := logx.Enter(ctx)
	var (
		x *consumer
		c *rabbitmq.Consumer
	)
	defer func() {
		if err != nil {
			log.Error(err)
		} else {
			e.AddConsumer(x)
			log.Info("sub created")
		}
		log.End()
	}()

	if e.closed.Load() || e.client == nil {
		return nil, codex.New(ECODE__CLI_CLOSED)
	}

	opt := e.Option.SubOption(options...)
	c, err = rabbitmq.NewConsumer(e.client, opt.queue, opt.options...)
	if err != nil {
		return nil, err
	}

	x = &consumer{
		sub:        c,
		cli:        e,
		log:        logx.NewStd().With("queue", opt.queue),
		callback:   opt.callback,
		autoAck:    !opt.disableAutoAck,
		mode:       opt.mode,
		worker:     opt.worker,
		hasher:     opt.hasher,
		bufferSize: opt.bufferSize,
	}

	return x, nil
}

func (e *Endpoint) Close() error {
	log := logx.From(context.Background())
	if e.closed.CompareAndSwap(false, true) {
		if e.client != nil {
			if err := e.client.Close(); err != nil {
				log.Error(fmt.Errorf("[driver:rabbit]failed to close client: %w", err))
			}
		}
		return e.ResourceManager.Close()
	}
	return nil
}

var (
	With  = mq.With[ProducerMessage, ConsumerMessage]
	From  = mq.From[ProducerMessage, ConsumerMessage]
	Must  = mq.Must[ProducerMessage, ConsumerMessage]
	Carry = mq.Carry[ProducerMessage, ConsumerMessage]
)
