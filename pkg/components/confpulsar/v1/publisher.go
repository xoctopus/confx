package confpulsar

import (
	"context"
	"sync/atomic"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/confx/pkg/components/confmq"
)

type publisher struct {
	cli      *Endpoint
	closed   atomic.Bool
	pub      pulsar.Producer
	sync     bool
	callback func(message confmq.Message, err error)
}

func (p *publisher) Topic() string {
	return p.pub.Topic()
}

func (p *publisher) Publish(ctx context.Context, v any) (err error) {
	_, log := logx.Enter(ctx)
	defer func() {
		if err != nil {
			log.Error(err)
		} else {
			log.Info("published")
		}
		log.End()
	}()

	if p.closed.Load() {
		return codex.New(ECODE__PUBLISHER_CLOSED)
	}
	if p.cli.closed.Load() {
		return codex.New(ECODE__CLIENT_CLOSED)
	}

	msg := confmq.NewMessage(ctx, p.pub.Topic(), v)
	log = log.With("topic", msg.Topic(), "msg_id", msg.ID())
	return p.PublishMessage(ctx, msg)
}

func (p *publisher) PublishMessage(ctx context.Context, msg confmq.Message) (err error) {
	_, log := logx.Enter(ctx, "topic", msg.Topic(), "msg_id", msg.ID())
	defer func() {
		if err != nil {
			log.Error(err)
		} else {
			log.Info("published")
		}
		log.End()
	}()

	data, err := msg.(confmq.MessageArshaler).Marshal()
	if err != nil {
		return err
	}
	raw := &pulsar.ProducerMessage{Payload: data}

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
					p.callback(msg, err)
				}
			}
		},
	)
	return nil
}

func (p *publisher) Close() {
	p.cli.ClosePublisher(p)
}
