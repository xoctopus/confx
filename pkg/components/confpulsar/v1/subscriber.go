package confpulsar

import (
	"context"
	"fmt"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/codex"

	. "github.com/xoctopus/confx/pkg/components/confmq"
)

type subscriber struct {
	topic    string
	cli      pulsar.Consumer
	cancel   context.CancelFunc
	failover func(context.Context, error)
}

func (s *subscriber) Topic() string {
	return s.topic
}

// Run starts consuming messages and processing them.
// Semantics:
//   - messages are ACKed immediately once received (at-most-once)
//   - returned error channel is unbuffered, caller MUST handle this error signal
//   - any error component-internal will terminate the Run loop
func (s *subscriber) Run(ctx context.Context, h func(context.Context, Message) error) <-chan error {
	ch := make(chan error)
	ctx, s.cancel = context.WithCancel(ctx)

	go func() {
		defer close(ch)

		for {
			msg, err := s.cli.Receive(ctx) // block call
			if err != nil {
				ch <- err
				return
			}

			if err = s.handle(ctx, msg, h); err != nil {
				if codex.IsCode(err, ECODE__PARSE_MESSAGE) ||
					codex.IsCode(err, ECODE__HANDLER_PANICKED) {
					if s.failover != nil {
						s.failover(ctx, err)
					}
					ch <- err
					return
				}
				s.cli.Nack(msg)
				continue
			}
			if err = s.cli.Ack(msg); err != nil {
				// TODO if ack failed
				logx.From(ctx).With(
					"msg_topic", msg.Topic(),
					"pulsar_msg_id", msg.ID().String(),
				).Warn(fmt.Errorf("failed ack message: %w", err))
			}
		}
	}()

	return ch
}

// handle wrapped consumer handle task
func (s *subscriber) handle(ctx context.Context, msg pulsar.Message, h func(context.Context, Message) error) (err error) {
	_, log := logx.From(ctx).Start(ctx, "subscriber.Handle",
		"sub_topic", s.topic,
		"msg_topic", msg.Topic(),
		"pulsar_msg_id", msg.ID().String(),
		"msg_pub_time", msg.PublishTime(),
	)
	defer func() {
		if r := recover(); r != nil {
			err = codex.Wrap(ECODE__HANDLER_PANICKED, fmt.Errorf("consumer handler panicked: %v", r))
		}
		if err != nil {
			log.Error(err)
		} else {
			log.Info("handled")
		}
		log.End()
	}()

	data := msg.Payload()
	m, err := ParseMessage(data)
	if err != nil {
		err = codex.Wrap(ECODE__PARSE_MESSAGE, err)
		return err
	}
	log = log.With("msg_id", m.ID())

	return h(ctx, m)
}

func (s *subscriber) Close() error {
	s.cancel()
	s.cli.Close()
	return nil
}
