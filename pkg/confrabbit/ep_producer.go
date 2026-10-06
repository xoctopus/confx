package confrabbit

import (
	"container/list"
	"context"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
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

	if p.sync {
		return p.publish(ctx, topic, raw)
	}

	// The async publish outlives this call, so it must not inherit the caller's
	// cancellation (a request context is typically canceled right after the
	// handler returns). Only context values (trace, logger, ...) are kept; the
	// publish timeout is derived inside publish and owned by the goroutine.
	ctx = context.WithoutCancel(ctx)
	go func() {
		err := p.publish(ctx, topic, raw)
		p.log.With("pub_at", msg.PublishedAt(), "result", err).Info("callback called")
		if p.callback != nil {
			p.callback(msg, err)
		}
	}()
	return nil
}

// publish writes raw to the broker and, when the publisher is in confirm mode,
// waits for the broker acknowledgement. p.timeout bounds the whole round trip.
//
// The AMQP write itself is not context-aware: once started it cannot be
// interrupted, and the frame may still reach the broker after a timeout has
// been reported. A timeout therefore means "not acknowledged in time", not
// "not delivered".
func (p *producer) publish(ctx context.Context, topic string, raw *amqp.Publishing) error {
	if p.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.timeout)
		defer cancel()
	}

	type result struct {
		confs rabbitmq.PublisherConfirmation
		err   error
	}
	// buffered so the writer never blocks once the timer has won the race
	results := make(chan result, 1)
	go func() {
		confs, err := p.pub.PublishWithDeferredConfirmWithContext(
			ctx,
			raw.Body,
			[]string{topic},
			rabbitmq.WithPublishOptionsExchange(p.exchange),
			rabbitmq.WithPublishOptionsHeaders(rabbitmq.Table(raw.Headers)),
			rabbitmq.WithPublishOptionsTimestamp(raw.Timestamp),
		)
		results <- result{confs: confs, err: err}
	}()

	var confs rabbitmq.PublisherConfirmation
	select {
	case <-ctx.Done():
		return WrapPublishTimeoutError(ctx.Err())
	case r := <-results:
		if r.err != nil {
			return WrapPublishTimeoutError(r.err)
		}
		confs = r.confs
	}

	for _, c := range confs {
		// nil when the publisher channel is not in confirm mode
		if c == nil {
			continue
		}
		acked, err := c.WaitContext(ctx)
		if err != nil {
			return WrapPublishTimeoutError(err)
		}
		if !acked {
			return codex.New(ECODE__PUB_NACKED)
		}
	}
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
