package confrabbit

import (
	"container/list"
	"context"
	"sync/atomic"
	"time"

	"github.com/wagslane/go-rabbitmq"
	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/confx/pkg/types/mq"
)

type producer struct {
	cli    *Endpoint
	elem   *list.Element
	closed atomic.Bool

	log      logx.Logger
	pub      *rabbitmq.Publisher
	topic    string
	exchange string
	timeout  time.Duration
	sync     bool
	callback mq.AsyncPubCallback[ProducerMessage]
}

func (p *producer) Topic() string {
	return p.topic
}

func (p *producer) Publish(ctx context.Context, topic string, payload []byte) (ProducerMessage, error) {
	msg := NewProducerMessage(topic, payload)
	return msg, p.PublishMessage(ctx, msg)
}

func (p *producer) PublishWithKey(ctx context.Context, topic, key string, payload []byte) (ProducerMessage, error) {
	msg := NewProducerMessage(topic, payload)
	msg.SetPartitionKey(key)
	return msg, p.PublishMessage(ctx, msg)
}

func (p *producer) PublishMessage(ctx context.Context, msg ProducerMessage) (err error) {
	if p.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.timeout)
		defer cancel()
	}

	_, log := logx.Enter(ctx)
	defer func() {
		if err != nil {
			log.Error(err)
		} else {
			log.Info("published")
		}
		log.End()
	}()

	topic := msg.Topic()
	if topic != p.topic {
		return codex.Errorf(
			ECODE__PUB_INVALID_MESSAGE,
			"unexpected topic: expect `%s` but got `%s`",
			p.topic, topic,
		)
	}
	log = log.With("topic", topic)

	if p.cli.closed.Load() {
		return codex.New(ECODE__CLI_CLOSED)
	}

	if p.closed.Load() {
		return codex.New(ECODE__PUB_CLOSED)
	}

	msg.RefreshPublishedAt()
	log = log.With("pub_at", msg.PublishedAt())
	raw := msg.Underlying()

	// RabbitMQ doesn't have an async publish that we can easily hook into like pulsar in wagslane/go-rabbitmq
	// It relies on publisher confirms under the hood if enabled.
	// We'll treat Publish as blocking here, or fire async if requested.

	if p.sync {
		err = p.pub.PublishWithContext(
			ctx,
			raw.Body,
			[]string{topic},
			rabbitmq.WithPublishOptionsExchange(p.exchange),
			rabbitmq.WithPublishOptionsHeaders(rabbitmq.Table(raw.Headers)),
			rabbitmq.WithPublishOptionsTimestamp(raw.Timestamp),
		)
		return
	}

	var done atomic.Bool
	go func() {
		err := p.pub.PublishWithContext(
			ctx,
			raw.Body,
			[]string{topic},
			rabbitmq.WithPublishOptionsExchange(p.exchange),
			rabbitmq.WithPublishOptionsHeaders(rabbitmq.Table(raw.Headers)),
			rabbitmq.WithPublishOptionsTimestamp(raw.Timestamp),
		)
		if done.CompareAndSwap(false, true) {
			p.log.With("pub_at", msg.PublishedAt(), "result", err).Info("callback called")
			if p.callback != nil {
				p.callback(msg, err)
			}
		}
	}()
	return nil
}

func (p *producer) Elem() *list.Element {
	return p.elem
}

func (p *producer) SetElem(elem *list.Element) {
	p.elem = elem
}

func (p *producer) Release(_ ...mq.ReleaseOptionFunc) error {
	if p.closed.CompareAndSwap(false, true) {
		p.pub.Close()
	}
	return nil
}

func (p *producer) Close() error {
	return p.cli.CloseProducer(p)
}
