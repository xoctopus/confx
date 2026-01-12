package confpulsar

import (
	"container/list"
	"context"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/codex"

	"github.com/xoctopus/confx/pkg/components/confmq"
)

type consumers struct {
	mtx sync.RWMutex
	lst list.List
}

func (c *consumers) add(s *subscriber) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	elem := c.lst.PushBack(s)
	s.elem = elem
}

func (c *consumers) close() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	for elem := c.lst.Front(); elem != nil; elem = elem.Next() {
		s := elem.Value.(*subscriber)
		s.cancel(codex.New(ECODE__SUBSCRIPTION_CANCELED))
		s.sub.Close()
		c.lst.Remove(elem)
	}
}

type subscriber struct {
	cli      *Endpoint
	elem     *list.Element
	topic    string
	sub      pulsar.Consumer
	cancel   context.CancelCauseFunc
	callback func(pulsar.Consumer, pulsar.Message, confmq.Message, error)
	autoAck  bool
}

func (s *subscriber) Topic() string {
	return s.topic
}

func (s *subscriber) Consumer() any {
	return s.sub
}

// Run starts consuming messages and processing them.
func (s *subscriber) Run(ctx context.Context, h func(context.Context, confmq.Message) error) <-chan error {
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
			if s.autoAck {
				if err = s.sub.Ack(msg); err != nil {
					log.With("action", "ack").Error(err) // TODO if ack failed should MUST ACKed
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
func (s *subscriber) handle(ctx context.Context, msg pulsar.Message, h func(context.Context, confmq.Message) error) (err error) {
	_, log := logx.From(ctx).Start(ctx, "subscriber.Handle", "pub_timestamp", msg.PublishTime())

	var m confmq.Message

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
	m, err = confmq.ParseMessage(data)
	if err != nil {
		err = codex.Wrap(ECODE__PARSE_MESSAGE, err)
		return err
	}
	log = log.With("msg_id", m.ID(), "topic", m.Topic())

	return h(ctx, m)
}

func (s *subscriber) Close() {
	s.cli.CloseSubscriber(s)
}
