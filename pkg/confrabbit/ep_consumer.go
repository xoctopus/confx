package confrabbit

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/wagslane/go-rabbitmq"
	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/codex"
	"github.com/xoctopus/x/slicex"

	"github.com/xoctopus/confx/pkg/types/mq"
)

type consumer struct {
	cli    *Endpoint
	elem   *list.Element
	closed atomic.Bool
	booted atomic.Bool
	sub    *rabbitmq.Consumer
	log    logx.Logger

	mode       mq.ConsumeHandleMode
	worker     uint16
	bufferSize uint16
	tasks      []chan rabbitmq.Delivery
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
	log.Info("processing started")
	for {
		select {
		case <-ctx.Done():
			err := errors.Join(slicex.Unique([]error{ctx.Err(), context.Cause(ctx)})...)
			log.Error(fmt.Errorf("processing stopped: %v", err))
			return err
		case m, ok := <-s.tasks[wid]:
			if !ok {
				return nil
			}
			logd := log.With("topic", m.RoutingKey)
			msg := NewConsumerMessage(m.Delivery)
			err := s.handle(ctx, msg)
			if err != nil {
				logd.With("action", "handle").Error(err)
			}

			if s.autoAck || s.callback == nil {
				if err == nil {
					if ackErr := m.Ack(false); ackErr != nil {
						logd.With("action", "ack").Error(ackErr)
					}
				} else {
					if ackErr := m.Nack(false, true); ackErr != nil {
						logd.With("action", "nack").Error(ackErr)
					}
				}
			}
		}
	}
}

func (s *consumer) run(ctx context.Context) error {
	err := s.sub.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
		var wid uint16
		switch s.mode {
		case mq.PartitionOrdered:
			wid = s.hasher(d.RoutingKey) % s.worker
		case mq.Concurrent:
			wid = uint16(d.DeliveryTag % uint64(s.worker))
		default:
			wid = 0
		}
		s.log.With("worker_id", wid, "routing_key", d.RoutingKey).Info("dispatched")

		if s.worker == 0 {
			msg := NewConsumerMessage(d.Delivery)
			err := s.handle(ctx, msg)
			if err != nil {
				return rabbitmq.NackRequeue
			}
			return rabbitmq.Ack
		}

		s.tasks[wid] <- d
		return rabbitmq.Manual
	})
	return err
}

func (s *consumer) dispatch(ctx context.Context) error {
	return s.run(ctx)
}

func (s *consumer) Run(ctx context.Context, h mq.SubHandler[ConsumerMessage]) error {
	if !s.booted.CompareAndSwap(false, true) {
		return codex.Errorf(ECODE__SUB_BOOTED, "reentered")
	}
	if s.cli.closed.Load() {
		return codex.New(ECODE__CLI_CLOSED)
	}
	if s.closed.Load() {
		return codex.New(ECODE__SUB_CLOSED)
	}

	if s.worker > 0 {
		s.tasks = make([]chan rabbitmq.Delivery, s.worker)
		for i := range s.tasks {
			s.tasks[i] = make(chan rabbitmq.Delivery, s.bufferSize)
		}
	}

	s.handler = h
	ctx, s.cancel = context.WithCancelCause(ctx)

	defer func() { _ = s.Close() }()

	if s.worker > 0 {
		for i := uint16(0); i < s.worker; i++ {
			go func(wid uint16) {
				_ = s.process(ctx, wid)
			}(i)
		}
	}

	log := s.log.With("workers", s.worker)
	log.Info("dispatching started")
	err := s.dispatch(ctx)
	log.Error(fmt.Errorf("dispatching stopped caused by %v", err))
	return err
}

func (s *consumer) handle(ctx context.Context, msg ConsumerMessage) (err error) {
	_, log := logx.Enter(
		ctx,
		"topic", msg.Topic(),
		"pub_at", msg.PublishedAt(),
		"latency", msg.Latency().Milliseconds(),
	)

	defer func() {
		if r := recover(); r != nil {
			err = codex.Errorf(ECODE__SUB_HANDLER_PANICKED, "cause: %v", r)
			if x, ok := r.(error); ok {
				err = codex.Wrap(ECODE__SUB_HANDLER_PANICKED, x)
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
	return m.Underlying().Ack(false)
}

func (s *consumer) Nack(m ConsumerMessage) error {
	return m.Underlying().Nack(false, true)
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
		cause := ECODE__SUB_CLOSED
		if opt.Unsub {
			cause = ECODE__SUB_UNSUBSCRIBED
		}
		if s.cancel != nil {
			s.cancel(codex.New(cause))
		}
		s.wg.Wait()
		if s.sub != nil {
			closed = true
			s.sub.Close()
		}
		if s.worker > 0 {
			for i := range s.tasks {
				close(s.tasks[i])
			}
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
