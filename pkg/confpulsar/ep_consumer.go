package confpulsar

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"

	"cgtech.gitlab.com/saitox/logx"
	"cgtech.gitlab.com/saitox/x/codex"
	"github.com/apache/pulsar-client-go/pulsar"

	"cgtech.gitlab.com/saitox/confx/pkg/types/mq"
)

type consumer struct {
	cli    *Endpoint
	elem   *list.Element
	closed atomic.Bool
	booted atomic.Bool
	sub    pulsar.Consumer
	log    logx.Logger

	mode       mq.ConsumeHandleMode
	worker     uint16
	bufferSize uint16
	tasks      []chan pulsar.Message
	wg         sync.WaitGroup

	hasher   mq.Hasher
	handler  mq.SubHandler[ConsumerMessage]
	callback mq.SubCallback[ConsumerMessage]

	cancel  context.CancelCauseFunc
	autoAck bool
}

var _ mq.Consumer[ConsumerMessage] = (*consumer)(nil)

func (s *consumer) process(ctx context.Context, wid uint16) error {
	s.wg.Add(1)
	defer s.wg.Done()

	log := s.log.With("worker_id", wid)
	log.Info("processing stated")
	for {
		select {
		case <-ctx.Done():
			err := errors.Join(ctx.Err(), context.Cause(ctx))
			log.Error(fmt.Errorf("processing stopped caused by %w", err))
			return err
		case m := <-s.tasks[wid]:
			logd := log.With("topic", m.Topic())
			msg := NewConsumerMessage(m)
			if err := s.handle(ctx, msg); err != nil {
				logd.With("action", "handle").Error(err)
			}
			if s.autoAck || s.callback == nil {
				if err := s.sub.Ack(m); err != nil {
					logd.With("action", "ack").Error(err)
				}
			}
		}
	}
}

func (s *consumer) dispatch(ctx context.Context) error {
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
		s.log.With("worker_id", wid, "order_key", msg.Key()).Info("dispatched")
		s.tasks[wid] <- msg
	}
}

// Run starts consuming messages and processing them.
func (s *consumer) Run(ctx context.Context, h mq.SubHandler[ConsumerMessage]) error {
	if !s.booted.CompareAndSwap(false, true) {
		return codex.Errorf(ERROR__SUB_BOOTED, "reentered")
	}
	if s.cli.closed.Load() {
		return codex.New(ERROR__CLI_CLOSED)
	}
	if s.closed.Load() {
		return codex.New(ERROR__SUB_CLOSED)
	}

	s.tasks = make([]chan pulsar.Message, s.worker)
	for i := range s.tasks {
		s.tasks[i] = make(chan pulsar.Message, s.bufferSize)
	}

	s.handler = h
	ctx, s.cancel = context.WithCancelCause(ctx)

	defer func() { _ = s.Close() }()
	for i := range s.worker {
		go func() {
			err := s.process(ctx, i)
			s.log.With("worker_id", i).Error(fmt.Errorf("processing stopped caused by: %w", err))
		}()
	}

	log := s.log.With("workers", s.worker)
	log.Info("dispatching started")
	err := s.dispatch(ctx)
	log.Error(fmt.Errorf("dispatching stopped caused by: %w", err))
	return err
}

// handle wrapped consumer handle task
func (s *consumer) handle(ctx context.Context, msg ConsumerMessage) (err error) {
	_, log := logx.Enter(
		ctx,
		"topic", msg.Topic(),
		"pub_at", msg.PublishedAt(),
		"latency", msg.Latency().Milliseconds(),
	)

	defer func() {
		if r := recover(); r != nil {
			err = codex.Errorf(ERROR__SUB_HANDLER_PANICKED, "cause: %v", r)
			if x, ok := r.(error); ok {
				err = codex.Wrap(ERROR__SUB_HANDLER_PANICKED, x)
			}
		}
		if err != nil {
			log.Error(err)
		} else {
			log.Info("handled")
		}
		if s.callback != nil {
			s.callback(s, msg, err)
		}
		log.End()
	}()
	return s.handler(ctx, msg)
}

func (s *consumer) Ack(m ConsumerMessage) error {
	return s.sub.Ack(m.Underlying())
}

func (s *consumer) Nack(m ConsumerMessage) error {
	s.sub.Nack(m.Underlying())
	return nil
}

func (s *consumer) Elem() *list.Element {
	return s.elem
}

func (s *consumer) SetElem(elem *list.Element) {
	s.elem = elem
}

func (s *consumer) Release(appliers ...mq.ReleaseOptionFunc) error {
	var (
		err    error
		opt    mq.ReleaseOption
		closed bool
	)

	for _, applier := range appliers {
		applier(&opt)
	}

	log := s.log.With("unsub", opt.Unsub)
	defer func() {
		if err != nil {
			log.Warn(err)
		}
		if closed {
			log.Info("consumer underlying closed")
		}
	}()

	if s.closed.CompareAndSwap(false, true) {
		cause := ERROR__SUB_CLOSED
		if opt.Unsub {
			cause = ERROR__SUB_UNSUBSCRIBED
		}
		if s.cancel != nil {
			s.cancel(codex.New(cause))
		}
		s.wg.Wait()
		if s.sub != nil {
			closed = true
			if opt.Unsub {
				err = s.sub.Unsubscribe()
			}
			s.sub.Close()
		}
		for i := range s.tasks {
			close(s.tasks[i])
		}
	}
	return err
}

func (s *consumer) Unsubscribe() error {
	return s.cli.ResourceManager.Unsubscribe(s)
}

func (s *consumer) Close() error {
	return s.cli.ResourceManager.CloseConsumer(s)
}
