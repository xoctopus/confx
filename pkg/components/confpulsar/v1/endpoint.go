// Package confpulsar defines component of redis client
// +genx:doc
package confpulsar

import (
	"context"
	"errors"
	"net/url"
	"sync/atomic"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/syncx"

	. "github.com/xoctopus/confx/pkg/components/confmq"
	"github.com/xoctopus/confx/pkg/types"
)

// Endpoint pulsar component endpoint
type Endpoint struct {
	types.Endpoint[PulsarOption]

	client    pulsar.Client
	producers syncx.Map[string, pulsar.Producer]
	closed    atomic.Bool
}

func (e *Endpoint) SetDefault() {
	if e.Endpoint.Address == "" {
		e.Endpoint.Address = "pulsar://localhost:6650"
	}
	if e.producers == nil {
		e.producers = syncx.NewXmap[string, pulsar.Producer]()
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
		v.Message = "endpoint is closed"
		return
	}

	opt := e.Option.PubOption("liveness")
	span := types.Cost()
	_, err := e.producer(ctx, &opt.options)
	v.TTL = types.Duration(span())
	if err != nil {
		v.Message = err.Error()
		return
	}
	v.Reachable = true
	return
}

func (e *Endpoint) producer(ctx context.Context, opt *pulsar.ProducerOptions) (p pulsar.Producer, err error) {
	must.BeTrueF(opt.Topic != "", "producer topic is required but got empty")

	_, log := logx.Enter(ctx, "topic", opt.Topic)
	defer func() {
		if err != nil {
			log.Error(err)
		} else {
			log.Info("")
		}
		log.End()
	}()

	if p, ok := e.producers.Load(opt.Topic); ok {
		log = log.With("producer", "loaded")
		return p, nil
	}

	log = log.With("producer", "create")
	producer, err := e.client.CreateProducer(*opt)
	if err != nil {
		return nil, err
	}

	actual, _ := e.producers.LoadOrStore(opt.Topic, producer)
	return actual, nil
}

func (e *Endpoint) Subscribe(ctx context.Context, topic string, options ...OptionApplier) (sub Subscriber, err error) {
	must.BeTrueF(topic != "", "consumer topic is required but got empty")

	_, log := logx.Enter(ctx, "topic", topic)
	defer func() {
		if err != nil {
			log.Error(err)
		} else {
			log.Info("sub created")
		}
		log.End()
	}()

	if e.closed.Load() || e.client == nil {
		return nil, errors.New("endpoint is closed")
	}

	opt := e.Option.SubOption(topic, options...)
	if opt.options.Type != pulsar.Exclusive {
		panic("any")
	}
	c, err := e.client.Subscribe(opt.options)
	if err != nil {
		return nil, err
	}

	return &subscriber{
		topic: topic,
		cli:   c,
	}, nil
}

func (e *Endpoint) Publish(ctx context.Context, msg Message, appliers ...OptionApplier) (err error) {
	must.BeTrueF(msg.Topic() != "", "publish topic is required but got empty")

	_, log := logx.Enter(ctx, "topic", msg.Topic(), "message_id", msg.ID())
	defer func() {
		if err != nil {
			log.Error(err)
		} else {
			log.Info("published")
		}
		log.End()
	}()

	if e.closed.Load() || e.client == nil {
		return errors.New("endpoint is closed")
	}

	var data []byte
	data, err = msg.(MessageArshaler).Marshal()
	if err != nil {
		return err
	}

	var (
		pub pulsar.Producer
		raw = &pulsar.ProducerMessage{Payload: data}
		opt = e.Option.PubOption(msg.Topic(), appliers...)
	)

	pub, err = e.producer(ctx, &opt.options)
	if err != nil {
		return err
	}

	if opt.sync {
		_, err = pub.Send(ctx, raw)
		return err
	}

	var done atomic.Bool
	pub.SendAsync(
		ctx, raw,
		func(_ pulsar.MessageID, _ *pulsar.ProducerMessage, err error) {
			if done.CompareAndSwap(false, true) {
				if opt.failover != nil {
					opt.failover(msg, err)
				}
			}
		},
	)
	return nil
}

func (e *Endpoint) Close() error {
	if e.closed.CompareAndSwap(false, true) {
		if e.producers != nil {
			for _, p := range e.producers.Range {
				p.Close()
			}
		}
		if e.client != nil {
			e.client.Close()
		}
	}
	return nil
}
