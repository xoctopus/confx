// Package confpulsar defines component of pulsar
// +genx:doc
package confpulsar

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/google/uuid"
	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/confx/pkg/types"
	"github.com/xoctopus/confx/pkg/types/liveness"
	"github.com/xoctopus/confx/pkg/types/mq"
)

// Endpoint pulsar component endpoint
type Endpoint struct {
	types.Endpoint[Option]

	client pulsar.Client
	closed atomic.Bool

	mq.ResourceManager
}

var _ mq.PubSub[ProducerMessage, ConsumerMessage] = (*Endpoint)(nil)

func (e *Endpoint) SetDefault() {
	if e.Endpoint.Address == "" {
		e.Endpoint.Address = "pulsar://localhost:6650"
	}

	if e.ResourceManager == nil {
		e.ResourceManager = mq.NewResourceManager()
	}

	e.Option.SetDefault()
	e.closed.Store(false)
}

func (e *Endpoint) Init(ctx context.Context) (err error) {
	if err = e.Endpoint.Init(); err != nil {
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
		e.client, err = pulsar.NewClient(opt)
		if err != nil {
			return codex.Wrap(ERROR__CLI_INIT_ERROR, err)
		}
	}

	return e.LivenessCheck(ctx).FailureReason()
}

// LivenessCheck helps to probe liveness of broker.
// Note: for avoiding backlogs. devs should config message ttl for `liveness` topic
// eg: pulsarctl topics set-message-ttl persistent://public/default/liveness -t 30
func (e *Endpoint) LivenessCheck(ctx context.Context) (v liveness.Result) {
	v = liveness.NewLivenessData()
	v.Start()

	if e.closed.Load() || e.client == nil {
		v.End(codex.New(ERROR__CLI_CLOSED))
		return
	}

	var (
		err  error
		r    pulsar.Reader
		p    mq.Producer[ProducerMessage]
		body = []byte(uuid.NewString())
	)
	defer func() {
		v.End(err)
	}()

	topic := "liveness"
	e.Option.PatchTopic(&topic)
	if strings.HasPrefix(topic, PERSISTENT) {
		topic = strings.Replace(topic, PERSISTENT, NON_PERSISTENT, 1)
	}

	r, err = e.client.CreateReader(pulsar.ReaderOptions{
		Topic:          topic,
		StartMessageID: pulsar.LatestMessageID(),
	})
	if err != nil {
		return
	}
	defer r.Close()

	p, err = e.NewProducer(
		ctx,
		WithPubTopic(topic),
		WithSyncPublish(),
		WithPubAccessMode(pulsar.ProducerAccessModeShared),
	)
	if err != nil {
		return
	}
	defer func() { _ = p.Close() }()

	msg := NewProducerMessage(topic, body)
	msg.RefreshPublishedAt()

	if err = p.PublishMessage(ctx, msg); err != nil {
		return
	}

	var (
		timeout = time.Second * 10
		cause   = fmt.Errorf("echo timeout in 10 seconds")
		cancel  context.CancelFunc
		echo    pulsar.Message
	)
	ctx, cancel = context.WithTimeoutCause(ctx, timeout, cause)
	defer cancel()
	for {
		if echo, err = r.Next(ctx); err == nil {
			if bytes.Equal(echo.Payload(), body) {
				return
			}
			continue
		}
		return
	}
}

func (e *Endpoint) NewProducer(ctx context.Context, options ...mq.OptionApplier) (_ mq.Producer[ProducerMessage], err error) {
	var (
		_, log = logx.Enter(ctx)
		p      pulsar.Producer
		x      *producer
		opt    = e.Option.PubOption(options...)
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
		return nil, codex.New(ERROR__CLI_CLOSED)
	}

	log = log.With("topic", opt.options.Topic, "sync", opt.sync)
	p, err = e.client.CreateProducer(opt.options)
	if err != nil {
		return
	}

	x = &producer{
		cli:      e,
		pub:      p,
		log:      logx.NewStd().With("producer", p.Name(), "topic", p.Topic()),
		sync:     opt.sync,
		callback: opt.callback,
	}
	return x, nil
}

func (e *Endpoint) NewConsumer(ctx context.Context, options ...mq.OptionApplier) (_ mq.Consumer[ConsumerMessage], err error) {
	var (
		_, log = logx.Enter(ctx)
		x      *consumer
		c      pulsar.Consumer
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
		return nil, codex.New(ERROR__CLI_CLOSED)
	}

	opt := e.Option.SubOption(options...)
	c, err = e.client.Subscribe(opt.options)
	if err != nil {
		return nil, err
	}
	log = log.With("consumer", c.Name(), "subgroup", c.Subscription())

	x = &consumer{
		sub:        c,
		cli:        e,
		log:        logx.NewStd().With("consumer", c.Name(), "subgroup", c.Subscription()),
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
	if e.closed.CompareAndSwap(false, true) {
		if e.client != nil {
			e.client.Close()
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
