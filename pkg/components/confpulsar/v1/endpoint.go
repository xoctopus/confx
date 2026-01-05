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
	"github.com/xoctopus/x/textx"

	. "github.com/xoctopus/confx/pkg/components/confmq"
	"github.com/xoctopus/confx/pkg/components/conftls"
	"github.com/xoctopus/confx/pkg/types"
)

type Endpoint struct {
	types.Endpoint

	client    pulsar.Client
	producers syncx.Map[string, pulsar.Producer]
	opt       options
	closed    atomic.Bool
}

type options struct {
	OperationTimeout  types.Duration `url:",default=100ms"`
	ConnTimeout       types.Duration `url:",default=5s"`
	KeepAliveInterval types.Duration `url:",default=1h"`
	MaxConnector      int            `url:",default=10"`
	MaxPending        int            `url:",default=100"`
	MaxBatching       uint           `url:",default=100"`
	conftls.X509KeyPair
}

var _ PubSub = (*Endpoint)(nil)

func (e *Endpoint) SetDefault() {
	if e.Endpoint.IsZero() {
		e.Endpoint = *must.NoErrorV(types.ParseEndpoint("pulsar://localhost:6650"))
	}
	if e.producers == nil {
		e.producers = syncx.NewXmap[string, pulsar.Producer]()
	}
	e.closed.Store(false)
}

func (e *Endpoint) Init(ctx context.Context) error {
	if err := textx.UnmarshalURL(e.Param, &e.opt); err != nil {
		return err
	}
	opt := pulsar.ClientOptions{
		URL:                     e.Endpoint.String(),
		ConnectionTimeout:       time.Duration(e.opt.ConnTimeout),
		OperationTimeout:        time.Duration(e.opt.OperationTimeout),
		KeepAliveInterval:       time.Duration(e.opt.KeepAliveInterval),
		MaxConnectionsPerBroker: e.opt.MaxConnector,
	}
	if !e.opt.X509KeyPair.IsZero() {
		opt.TLSConfig = e.opt.X509KeyPair.Config()
		opt.URL = (&url.URL{
			Scheme:   "pulsar+ssl",
			Host:     e.Hostname(),
			User:     url.UserPassword(e.Username, e.Password.String()),
			Path:     "/" + e.Base,
			RawQuery: e.Param.Encode(),
		}).String()
	}

	if e.client == nil {
		client, err := pulsar.NewClient(opt)
		if err != nil {
			return err
		}
		e.client = client
	}

	if r := e.LivenessCheck(ctx); !r[e].Reachable {
		return errors.New(r[e].Msg)
	}

	return nil
}

func (e *Endpoint) LivenessCheck(ctx context.Context) (r map[types.Component]types.LivenessCheckDetail) {
	d := types.LivenessCheckDetail{}

	defer func() {
		r = map[types.Component]types.LivenessCheckDetail{e: d}
	}()

	if e.closed.Load() || e.client == nil {
		d.Msg = "endpoint is closed"
		return
	}

	span := types.Cost()
	_, err := e.producer(ctx, &pulsar.ProducerOptions{Topic: "liveness"})
	d.TTL = span()
	if err != nil {
		d.Msg = err.Error()
		return
	}
	d.Reachable = true
	return
}

func (e *Endpoint) Options() url.Values {
	param, _ := textx.MarshalURL(&e.opt)
	return param
}

func (e *Endpoint) producer(ctx context.Context, opt *pulsar.ProducerOptions) (pulsar.Producer, error) {
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

func (e *Endpoint) Subscribe(ctx context.Context, topic string) (Subscriber, error) {
	must.BeTrueF(topic != "", "consumer topic is required but got empty")

	if e.closed.Load() || e.client == nil {
		return nil, errors.New("endpoint is closed")
	}

	opt := pulsar.ConsumerOptions{
		Topic:            topic,
		RetryEnable:      true,
		SubscriptionName: topic,
	}

	c, err := e.client.Subscribe(opt)
	if err != nil {
		return nil, err
	}

	return &subscriber{
		topic: topic,
		cli:   c,
	}, nil
}

func (e *Endpoint) Publish(ctx context.Context, msg Message, options ...OptionApplier) (err error) {
	must.BeTrueF(msg.Topic() != "", "publish topic is required but got empty")

	var opt = newDefaultProducerOption(e.opt, msg.Topic())
	for _, applier := range options {
		applier.Apply(opt)
	}

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
		for _, p := range e.producers.Range {
			p.Close()
		}
		if e.client != nil {
			e.client.Close()
		}
	}
	return nil
}
