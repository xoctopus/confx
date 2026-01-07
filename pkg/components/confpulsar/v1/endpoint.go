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
	e.closed.Store(false)
}

func (e *Endpoint) Init(ctx context.Context) error {
	if err := e.Endpoint.Init(); err != nil {
		return err
	}

	opt := pulsar.ClientOptions{
		URL:                     e.Endpoint.String(),
		ConnectionTimeout:       time.Duration(e.Option.ConnTimeout),
		OperationTimeout:        time.Duration(e.Option.OperationTimeout),
		KeepAliveInterval:       time.Duration(e.Option.KeepAliveInterval),
		MaxConnectionsPerBroker: e.Option.MaxConnector,
	}

	if !e.Endpoint.Cert.IsZero() {
		opt.TLSConfig = e.Endpoint.Cert.Config()
		u := e.URL()
		opt.URL = (&url.URL{
			Scheme: "pulsar+ssl",
			Host:   u.Host,
			User:   u.User,
			Path:   u.Path,
		}).String()
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

	opt := newPubOption(&e.Option, "liveness")
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

func (e *Endpoint) producer(_ context.Context, opt *pulsar.ProducerOptions) (pulsar.Producer, error) {
	must.BeTrueF(opt.Topic != "", "producer topic is required but got empty")

	if p, ok := e.producers.Load(opt.Topic); ok {
		return p, nil
	}

	producer, err := e.client.CreateProducer(*opt)
	if err != nil {
		return nil, err
	}

	actual, _ := e.producers.LoadOrStore(opt.Topic, producer)
	return actual, nil
}

func (e *Endpoint) Subscribe(_ context.Context, topic string, options ...OptionApplier) (Subscriber, error) {
	must.BeTrueF(topic != "", "consumer topic is required but got empty")

	if e.closed.Load() || e.client == nil {
		return nil, errors.New("endpoint is closed")
	}

	opt := newSubOption(&e.Option, topic, options...)
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
		opt = newPubOption(&e.Option, msg.Topic(), appliers...)
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
				if opt.callback != nil {
					opt.callback(msg, err)
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
