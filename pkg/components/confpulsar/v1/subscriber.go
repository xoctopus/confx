package confpulsar

import (
	"context"
	"fmt"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/xoctopus/logx"

	. "github.com/xoctopus/confx/pkg/components/confmq"
)

type subscriber struct {
	topic  string
	cli    pulsar.Consumer
	cancel context.CancelFunc
}

func (s *subscriber) Topic() string {
	return s.topic
}

// Run starts consuming messages and processing them.
// Semantics:
//   - messages are ACKed immediately once received (at-most-once)
//   - returned error channel is unbuffered, caller MUST handle this error signal
//   - any error component-internal will terminate the Run loop
func (s *subscriber) Run(ctx context.Context, h func(context.Context, Message)) <-chan error {
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
			_ = s.cli.Ack(msg)

			if err = s.handle(ctx, msg, h); err != nil {
				ch <- err
				return
			}
		}
	}()

	return ch
}

// handle wrapped consumer handle task
func (s *subscriber) handle(ctx context.Context, msg pulsar.Message, h func(context.Context, Message)) (err error) {
	_, log := logx.From(ctx).Start(
		ctx, "subscriber.Handle",
		"topic", s.topic,
		"pulsar_msg_id", msg.ID(),
	)
	defer log.End()

	defer func() {
		r := recover()
		if r == nil {
			return
		}
		err = fmt.Errorf("consumer handler panicked: %v", r)
		log.Warn(err)
	}()

	data := msg.Payload()
	m, err := ParseMessage(data)
	if err != nil {
		log.Warn(fmt.Errorf("failed to parse message: %w", err))
		return err
	}

	h(ctx, m)
	return nil
}

func (s *subscriber) Close() error {
	s.cancel()
	s.cli.Close()
	return nil
}
