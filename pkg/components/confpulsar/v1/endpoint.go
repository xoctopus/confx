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
	"github.com/xoctopus/x/syncx"

	"github.com/xoctopus/confx/pkg/components/confmq"
	"github.com/xoctopus/confx/pkg/types"
)

// Endpoint pulsar component endpoint
type Endpoint struct {
	types.Endpoint[Option]

	client    pulsar.Client
	closed    atomic.Bool
	producers syncx.Map[string, *publisher]
	consumers *consumers
}

func (e *Endpoint) SetDefault() {
	if e.Endpoint.Address == "" {
		e.Endpoint.Address = "pulsar://localhost:6650"
	}
	if e.producers == nil {
		e.producers = syncx.NewXmap[string, *publisher]()
	}
	if e.consumers == nil {
		e.consumers = &consumers{}
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
			return err
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

	span := types.Cost()
	p, err := e.Publisher(ctx, WithPubTopic("liveness"), WithSyncPublish())
	if err != nil {
		v.Message = err.Error()
		return
	}
	msg := confmq.NewMessage(ctx, "liveness", nil)
	if err = p.PublishMessage(ctx, msg); err != nil {
		v.Message = err.Error()
		return
	}

	s, err := e.Subscriber(ctx, WithSubTopic("liveness"))
	if err != nil {
		v.Message = err.Error()
		return
	}
	defer s.Close()

	select {
	case <-s.Run(ctx, func(ctx context.Context, m confmq.Message) error {
		if m.ID() == msg.ID() {
			v.TTL = types.Duration(span())
			v.Reachable = true
			s.Close()
		}
		return nil
	}):
	case <-time.After(time.Minute):
		v.TTL = types.Duration(span())
		v.Reachable = false
		v.Message = "liveness check timeout"
	}

	return
}

func (e *Endpoint) Publisher(ctx context.Context, options ...confmq.OptionApplier) (p confmq.Publisher, err error) {
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
		opt    = e.Option.PubOption(options...)
		pub    pulsar.Producer
		loaded bool
	)
	must.BeTrueF(opt.options.Topic != "", "producer topic is required")
	opt.options.Topic = e.Option.String() + opt.options.Topic

	log = log.With("topic", opt.options.Topic, "sync", opt.sync)
	p, loaded = e.producers.Load(opt.options.Topic)
	if loaded {
		log = log.With("producer", "loaded")
		return
	}

	pub, err = e.client.CreateProducer(opt.options)
	if err != nil {
		return
	}

	log = log.With("producer", "created")
	p, _ = e.producers.LoadOrStore(opt.options.Topic, &publisher{
		cli:      e,
		pub:      pub,
		sync:     opt.sync,
		callback: opt.callback,
	})
	return
}

func (e *Endpoint) Subscriber(ctx context.Context, options ...confmq.OptionApplier) (_ confmq.Subscriber, err error) {
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
	log = log.With("topic", opt.options.Topic, "name", opt.options.SubscriptionName)

	s, err := e.client.Subscribe(opt.options)
	if err != nil {
		return nil, err
	}

	sub := &subscriber{
		sub:      s,
		cli:      e,
		callback: opt.callback,
		autoAck:  !opt.disableAutoAck,
	}
	e.consumers.add(sub)
	return sub, nil
}

func (e *Endpoint) Close() error {
	if e.closed.CompareAndSwap(false, true) {
		if e.client != nil {
			e.client.Close()
		}
		if e.producers != nil {
			for _, p := range e.producers.Range {
				if p.closed.CompareAndSwap(false, true) {
					p.pub.Close()
				}
			}
		}
		if e.consumers != nil {
			e.consumers.close()
		}
	}
	return nil
}

func (e *Endpoint) CloseSubscriber(sub confmq.Subscriber) {
	if s, ok := sub.(*subscriber); ok {
		e.consumers.mtx.Lock()
		defer e.consumers.mtx.Unlock()
		e.consumers.lst.Remove(s.elem)
		s.cancel(codex.New(ECODE__SUBSCRIPTION_CANCELED))
		s.sub.Close()
	}
}

func (e *Endpoint) ClosePublisher(pub confmq.Publisher) {
	if p, ok := pub.(*publisher); ok {
		if x, _ := e.producers.LoadAndDelete(p.Topic()); x != nil {
			if x.closed.CompareAndSwap(false, true) {
				x.pub.Close()
			}
		}
	}
}
