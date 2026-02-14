// Package confpulsar defines component of redis client
// +genx:doc
package confpulsar

import (
	"context"
	"errors"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/codex"
	"github.com/xoctopus/x/misc/must"

	"github.com/xoctopus/confx/pkg/types"
	"github.com/xoctopus/confx/pkg/types/mq"
)

// Endpoint pulsar component endpoint
type Endpoint struct {
	types.Endpoint[Option]

	client    pulsar.Client
	closed    atomic.Bool
	producers *manager[*publisher]
	consumers *manager[*subscriber]
}

func (e *Endpoint) SetDefault() {
	if e.Endpoint.Address == "" {
		e.Endpoint.Address = "pulsar://localhost:6650"
	}
	if e.producers == nil {
		e.producers = &manager[*publisher]{}
	}
	if e.consumers == nil {
		e.consumers = &manager[*subscriber]{}
	}
	e.Option.SetDefault()
	e.closed.Store(false)
}

func (e *Endpoint) Init(ctx context.Context) error {
	if err := e.Endpoint.Init(); err != nil {
		return err
	}

	opt := e.Option.ClientOption(e.Endpoint.Endpoint())
	if !e.Endpoint.Cert.IsZero() {
		u := e.URL()
		opt.URL = (&url.URL{
			Scheme: "pulsar+ssl",
			Host:   u.Host,
			User:   u.User,
			Path:   u.Path,
		}).String()
		opt.TLSConfig = e.Endpoint.Cert.Config()
	}

	if e.client == nil {
		client, err := pulsar.NewClient(opt)
		if err != nil {
			return codex.Wrap(ECODE__FAILED_TO_INIT_CLIENT, err)
		}
		e.client = client
	}

	if d := e.LivenessCheck(ctx); !d.Reachable {
		return errors.New(d.Message)
	}

	return nil
}

func (e *Endpoint) LivenessCheck(ctx context.Context) (v types.LivenessData) {
	if e.closed.Load() || e.client == nil {
		v.Message = codex.New(ECODE__CLIENT_CLOSED).Error()
		return
	}

	var (
		echo = make(chan struct{}, 1)
		msg  = mq.NewMessage(ctx, "liveness", nil)
		span func() time.Duration
	)

	defer close(echo)

	s, err := e.Subscriber(
		ctx,
		WithSubTopic("liveness"),
		WithSubType(pulsar.Exclusive),
		WithSubCallback(func(c pulsar.Consumer, m pulsar.Message, x mq.Message, _ error) {
			if msg != nil && x != nil && x.ID() == msg.ID() {
				echo <- struct{}{}
			}
		}),
	)
	if err != nil {
		v.Message = err.Error()
		return
	}
	defer s.Close()

	p, err := e.Publisher(
		ctx,
		WithPubTopic("liveness"),
		WithSyncPublish(),
		WithPubAccessMode(pulsar.ProducerAccessModeExclusive),
	)
	if err != nil {
		v.Message = err.Error()
		return
	}
	defer p.Close()

	span = types.Cost()
	if err = p.PublishMessage(ctx, msg); err != nil {
		v.Message = err.Error()
		return
	}

	var (
		cancel    context.CancelFunc
		etimedout = errors.New("liveness check timeout")
	)
	ctx, cancel = context.WithTimeoutCause(ctx, time.Second<<2, etimedout)
	defer cancel()

	go func() {
		err = s.Run(ctx, func(context.Context, mq.Message) error { return nil })
	}()

	select {
	case <-echo:
		v.RTT = types.Duration(span())
		v.Reachable = true
	case <-ctx.Done():
		v.RTT = types.Duration(span())
		v.Reachable = false
		v.Message = errors.Join(ctx.Err(), context.Cause(ctx)).Error()
	}
	return
}

func (e *Endpoint) Publisher(ctx context.Context, options ...mq.OptionApplier) (_ mq.Publisher, err error) {
	_, log := logx.Enter(ctx)
	defer func() {
		if err != nil {
			log.Error(err)
		} else {
			log.Info("pub created")
		}
		log.End()
	}()

	if e.closed.Load() || e.client == nil {
		return nil, codex.New(ECODE__CLIENT_CLOSED)
	}

	var (
		opt = e.Option.PubOption(options...)
		p   pulsar.Producer
	)
	must.BeTrueF(opt.options.Topic != "", "producer topic is required")
	opt.options.Topic = e.Option.String() + opt.options.Topic

	log = log.With("topic", opt.options.Topic, "sync", opt.sync)
	p, err = e.client.CreateProducer(opt.options)
	if err != nil {
		return
	}

	log = log.With("producer", "created")
	pub := &publisher{
		cli:      e,
		pub:      p,
		sync:     opt.sync,
		callback: opt.callback,
	}
	e.producers.Add(pub)
	return pub, nil
}

func (e *Endpoint) Subscriber(ctx context.Context, options ...mq.OptionApplier) (_ mq.Subscriber, err error) {
	_, log := logx.Enter(ctx)
	defer func() {
		if err != nil {
			log.Error(err)
		} else {
			log.Info("sub created")
		}
		log.End()
	}()

	if e.closed.Load() || e.client == nil {
		return nil, codex.New(ECODE__CLIENT_CLOSED)
	}

	opt := e.Option.SubOption(options...)
	must.BeTrueF(
		opt.options.Topic != "" || len(opt.options.Topics) > 0 ||
			opt.options.TopicsPattern != "" || opt.options.SubscriptionName == "",
		"consumer topic is required",
	)
	opt.options.Topic = e.Option.String() + opt.options.Topic
	if opt.options.TopicsPattern != "" {
		opt.options.TopicsPattern = e.Option.String() + opt.options.TopicsPattern
	}
	for i := range opt.options.Topics {
		opt.options.Topics[i] = e.Option.String() + opt.options.Topics[i]
	}

	log = log.With("subscription", opt.options.SubscriptionName)
	s, err := e.client.Subscribe(opt.options)
	if err != nil {
		return nil, err
	}
	log = log.With("consumer", s.Name())

	sub := &subscriber{
		sub:        s,
		cli:        e,
		callback:   opt.callback,
		autoAck:    !opt.disableAutoAck,
		mode:       opt.mode,
		worker:     opt.worker,
		hasher:     opt.hasher,
		bufferSize: opt.bufferSize,
	}
	e.consumers.Add(sub)
	return sub, nil
}

func (e *Endpoint) Close() error {
	if e.closed.CompareAndSwap(false, true) {
		if e.client != nil {
			e.client.Close()
		}
		if e.producers != nil {
			e.producers.Close()
		}
		if e.consumers != nil {
			e.consumers.Close()
		}
	}
	return nil
}

func (e *Endpoint) CloseSubscriber(sub mq.Subscriber) {
	if s, ok := sub.(*subscriber); ok {
		e.consumers.Remove(s)
	}
}

func (e *Endpoint) ClosePublisher(pub mq.Publisher) {
	if p, ok := pub.(*publisher); ok {
		e.producers.Remove(p)
	}
}

func (e *Endpoint) ConsumerCount() int {
	return e.consumers.Len()
}

func (e *Endpoint) ProducerCount() int {
	return e.producers.Len()
}
