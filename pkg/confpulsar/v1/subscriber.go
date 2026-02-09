package confpulsar

import (
	"container/list"
	"context"
	"fmt"
	"runtime/debug"
	"sync/atomic"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/confx/pkg/types/mq"
)

type subscriber struct {
	cli    *Endpoint
	elem   *list.Element
	closed atomic.Bool
	sub    pulsar.Consumer

	cancel   context.CancelCauseFunc
	callback func(pulsar.Consumer, pulsar.Message, mq.Message, error)
	autoAck  bool
}

// Run starts consuming messages and processing them.
// NOTE:
//  1. The returned error channel is unbuffered. The caller MUST handle it.
//  2. If `autoAck` is not enabled and `callback` is configured, the subscriber
//     will NOT acknowledge messages automatically.
//  3. If a callback is configured, it will be invoked after message processed
func (s *subscriber) Run(ctx context.Context, h func(context.Context, mq.Message) error) <-chan error {
	ch := make(chan error)
	ctx, s.cancel = context.WithCancelCause(ctx)

	go func() {
		defer close(ch)

		for {
			log := logx.From(ctx)
			msg, err := s.sub.Receive(ctx) // block call until subscriber closed
			if err != nil {
				log.Error(err)
				ch <- err
				return
			}

			log = log.With("msg_topic", msg.Topic(), "pulsar_msg_id", msg.ID().String())
			if s.autoAck || s.callback == nil {
				if err = s.sub.Ack(msg); err != nil {
					log.With("action", "ack").Error(err)
				}
			}

			err = s.handle(ctx, msg, h)
			if err != nil {
				log.With("action", "handle").Error(err)
			}
		}
	}()

	return ch
}

// handle wrapped consumer handle task
func (s *subscriber) handle(ctx context.Context, msg pulsar.Message, h func(context.Context, mq.Message) error) (err error) {
	_, log := logx.From(ctx).Start(ctx, "subscriber.Handle", "pub_timestamp", msg.PublishTime())

	var m mq.Message

	defer func() {
		if r := recover(); r != nil {
			err = codex.Wrap(ECODE__HANDLER_PANICKED, fmt.Errorf("consumer handler panicked: %v\n%s", r, string(debug.Stack())))
		}
		if err != nil {
			log.Error(err)
		} else {
			log.Info("handled")
		}
		if s.callback != nil {
			s.callback(s.sub, msg, m, err)
		}
		log.End()
	}()

	data := msg.Payload()
	m, err = mq.ParseMessage(data)
	if err != nil {
		err = codex.Wrap(ECODE__PARSE_MESSAGE, err)
		return err
	}
	m.(mq.MutMessage).SetPubOrderedKey(msg.Key())
	log = log.With("msg_id", m.ID(), "topic", m.Topic())

	return h(ctx, m)
}

func (s *subscriber) Elem() *list.Element {
	return s.elem
}

func (s *subscriber) SetElem(elem *list.Element) {
	s.elem = elem
}

func (s *subscriber) close() {
	if s.closed.CompareAndSwap(false, true) {
		if s.cancel != nil {
			s.cancel(codex.New(ECODE__SUBSCRIPTION_CANCELED))
		}
		if s.sub != nil {
			s.sub.Close()
		}
	}
}

func (s *subscriber) Close() {
	s.cli.CloseSubscriber(s)
}
