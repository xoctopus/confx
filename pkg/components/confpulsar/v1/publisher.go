package confpulsar

import (
	"container/list"
	"context"
	"sync/atomic"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/confx/pkg/components/confmq"
)

type publisher struct {
	cli    *Endpoint
	elem   *list.Element
	closed atomic.Bool

	pub      pulsar.Producer
	sync     bool
	callback func(message confmq.Message, err error)
}

func (p *publisher) Topic() string {
	return p.pub.Topic()
}

func (p *publisher) publish(ctx context.Context, m confmq.Message) (err error) {
	_, log := logx.Enter(ctx, "topic", m.Topic(), "msg_id", m.ID())
	defer func() {
		if err != nil {
			log.Error(err)
		} else {
			log.Info("published")
		}
		log.End()
	}()

	if p.cli.closed.Load() {
		return codex.New(ECODE__CLIENT_CLOSED)
	}

	if p.closed.Load() {
		return codex.New(ECODE__PUBLISHER_CLOSED)
	}

	data, err := m.(confmq.MessageArshaler).Marshal()
	if err != nil {
		return err
	}

	raw := &pulsar.ProducerMessage{Payload: data}
	if x, ok := m.(confmq.OrderedMessage); ok {
		raw.Key = x.PubOrderedKey()
	}

	if p.sync {
		_, err = p.pub.Send(ctx, raw)
		return
	}

	var done atomic.Bool
	p.pub.SendAsync(
		ctx, raw,
		func(_ pulsar.MessageID, _ *pulsar.ProducerMessage, err error) {
			if done.CompareAndSwap(false, true) {
				if p.callback != nil {
					p.callback(m, err)
				}
			}
		},
	)
	return nil
}

func (p *publisher) Publish(ctx context.Context, v any) (err error) {
	return p.publish(ctx, confmq.NewMessage(ctx, p.pub.Topic(), v))
}

func (p *publisher) PublishMessage(ctx context.Context, msg confmq.Message) (err error) {
	if p.cli.Option.Topic(msg.Topic()) != p.pub.Topic() {
		return codex.Errorf(ECODE__PUB_INVALID_MESSAGE, "unexpected topic")
	}
	return p.publish(ctx, msg)
}

func (p *publisher) Elem() *list.Element {
	return p.elem
}

func (p *publisher) SetElem(elem *list.Element) {
	p.elem = elem
}

func (p *publisher) close() {
	if p.closed.CompareAndSwap(false, true) {
		p.pub.Close()
	}
}

func (p *publisher) Close() {
	p.cli.ClosePublisher(p)
}
