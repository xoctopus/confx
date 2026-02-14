package confpulsar

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"math"
	"runtime/debug"
	"sync"
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

	mode       mq.ConsumeMode
	worker     uint16
	hasher     mq.Hasher
	bufferSize uint16
	tasks      []chan pulsar.Message
	wg         sync.WaitGroup
	handler    mq.Handler
	booted     atomic.Bool

	cancel   context.CancelCauseFunc
	callback func(pulsar.Consumer, pulsar.Message, mq.Message, error)
	autoAck  bool
}

func (s *subscriber) process(ctx context.Context, wid uint16) error {
	s.wg.Add(1)
	defer s.wg.Done()

	log := logx.From(ctx).With("worker_id", wid, "consumer", s.sub.Name(), "subscription", s.sub.Subscription())
	log.Info("processing stated")
	for {
		select {
		case <-ctx.Done():
			err := errors.Join(ctx.Err(), context.Cause(ctx))
			log.Error(fmt.Errorf("processing stopped: %w", err))
			return err
		case msg := <-s.tasks[wid]:
			log2 := log.With("topic", msg.Topic())
			if err := s.handle(ctx, msg, s.handler); err != nil {
				log2.With("action", "handle").Error(err)
			}
			if s.autoAck || s.callback == nil {
				if err := s.sub.Ack(msg); err != nil {
					log2.With("action", "ack").Error(err)
				}
			}
		}
	}
}

func (s *subscriber) dispatch(ctx context.Context) error {
	var (
		count uint16
		wid   uint16
	)
	for {
		// block call until subscriber closed
		msg, err := s.sub.Receive(ctx)
		if err != nil {
			return errors.Join(err, context.Cause(ctx))
		}
		switch s.mode {
		case mq.PartitionOrdered:
			wid = s.hasher(msg.Key()) % s.worker
		case mq.Concurrent:
			count = (count + 1) % math.MaxUint16
			wid = count % s.worker
		default:
			wid = 0
		}
		logx.From(ctx).With("worker_id", wid, "ordered_key", msg.Key()).Info("dispatched")
		s.tasks[wid] <- msg
	}
}

// Run starts consuming messages and processing them.
// NOTE:
//  1. The returned error channel is unbuffered. The caller MUST handle it.
//  2. If `autoAck` is not enabled and `callback` is configured, the subscriber
//     will NOT acknowledge messages automatically.
//  3. If a callback is configured, it will be invoked after message processed
func (s *subscriber) Run(ctx context.Context, h func(context.Context, mq.Message) error) error {
	if !s.booted.CompareAndSwap(false, true) {
		return codex.Errorf(ECODE__SUBSCRIBER_BOOTED, "reentered")
	}

	defer s.Close()

	s.handler = h
	s.tasks = make([]chan pulsar.Message, s.worker)
	for i := range s.tasks {
		s.tasks[i] = make(chan pulsar.Message, s.bufferSize)
	}
	ctx, s.cancel = context.WithCancelCause(ctx)

	for i := range s.worker {
		go func() {
			err := s.process(ctx, i)
			logx.From(ctx).With("worker_id", i).Error(fmt.Errorf("processing stopped caused by: %w", err))
		}()
	}
	log := logx.From(ctx).With("worker_count", s.worker, "consumer", s.sub.Name(), "subscription", s.sub.Subscription())
	log.Info("dispatching started")
	err := s.dispatch(ctx)
	log.Error(fmt.Errorf("dispatching stopped caused by: %w", err))
	return err
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
		s.wg.Wait()
		if s.sub != nil {
			s.sub.Close()
			logx.From(context.Background()).With(
				"consumer", s.sub.Name(),
				"subscription", s.sub.Subscription(),
			).Info("underlying consumer closed")
		}
	}
}

func (s *subscriber) Close() {
	s.cli.CloseSubscriber(s)
}
