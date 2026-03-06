package confpulsar

import (
	"container/list"
	"context"
	"sync/atomic"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/confx/pkg/types/mq"
)

type producer struct {
	cli    *Endpoint
	elem   *list.Element
	closed atomic.Bool

	log      logx.Logger
	pub      pulsar.Producer
	sync     bool
	callback mq.AsyncPubCallback[ProducerMessage]
}

func (p *producer) Topic() string {
	return p.pub.Topic()
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
	if p.cli.Option.PatchTopic(&topic); topic != p.pub.Topic() {
		return codex.Errorf(
			ERROR__PUB_INVALID_MESSAGE,
			"unexpected topic: expect `%s` but got `%s`",
			p.pub.Topic(), topic,
		)
	}

	if p.cli.closed.Load() {
		return codex.New(ERROR__CLI_CLOSED)
	}

	if p.closed.Load() {
		return codex.New(ERROR__PUB_CLOSED)
	}

	msg.RefreshPublishedAt()
	log = log.With("last_sequence", p.pub.LastSequenceID(), "pub_at", msg.PublishedAt())
	raw := msg.Underlying()

	if p.sync {
		_, err = p.pub.Send(ctx, raw)
		return
	}

	var done atomic.Bool
	p.pub.SendAsync(
		ctx, raw,
		func(_ pulsar.MessageID, m *pulsar.ProducerMessage, err error) {
			if done.CompareAndSwap(false, true) {
				p.log.With("pub_at", msg.PublishedAt(), "result", err).Info("callback called")
				if p.callback != nil {
					p.callback(msg, err)
				}
			}
		},
	)
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
