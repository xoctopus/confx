package hack

import (
	"context"
	"net/url"
	"testing"
	"time"

	. "github.com/xoctopus/x/testx"

	. "github.com/xoctopus/confx/pkg/confpulsar"
	"github.com/xoctopus/confx/pkg/types/mq"
)

func WithPulsar(ctx context.Context, t testing.TB, dsn string) context.Context {
	Check(t)

	_, err := url.Parse(dsn)
	Expect(t, err, Succeed())

	ep := &Endpoint{}
	ep.Address = dsn
	ep.SetDefault()

	err = retrier.Do(func() error { return ep.Init(ctx) })
	Expect(t, err, Succeed())

	t.Cleanup(func() {
		_ = ep.Close()
	})
	return With(ctx, ep)
}

func TryWithPulsar(ctx context.Context, t testing.TB, dsn string) (context.Context, error) {
	Check(t)

	_, err := url.Parse(dsn)
	Expect(t, err, Succeed())

	ep := &Endpoint{}
	ep.Address = dsn
	ep.SetDefault()

	err = ep.Init(ctx)
	if err == nil {
		t.Cleanup(func() { _ = ep.Close() })
		return With(ctx, ep), nil
	}
	return ctx, err
}

func WithPulsarLost(ctx context.Context, t testing.TB, dsn string) context.Context {
	_, err := url.Parse(dsn)
	Expect(t, err, Succeed())

	ep := &Endpoint{}
	ep.SetDefault()
	ep.Address = dsn

	Expect(t, ep.Init(ctx), Failed())
	return Carry(ep)(ctx)
}

func RunPulsarPubSubTestSuite(
	ctx context.Context,
	t testing.TB,
	dsn string,
	messages []ProducerMessage, // messages published for testing
	termsig chan struct{}, // terminal subscribing
	handler func(context.Context, ConsumerMessage) error,
	timeout time.Duration, // test case throttling
	unsub bool, // if true unsubscribe
	appliers ...mq.OptionApplier,
) {
	ctx = WithPulsar(ctx, t, dsn)
	ps := Must(ctx)

	// factory producer and consumer for testing
	pub, err := ps.NewProducer(ctx, appliers...)
	Expect(t, err, Succeed())
	sub, err := ps.NewConsumer(ctx, appliers...)
	Expect(t, err, Succeed())

	// consuming
	go func() {
		if nil != sub.Run(ctx, handler) {
			return
		}
	}()

	// producing
	for _, m := range messages {
		Expect(t, pub.PublishMessage(ctx, m), Succeed())
	}

	t.Cleanup(func() {
		_ = pub.Close()
		if unsub {
			_ = sub.(mq.Unsubscriber).Unsubscribe()
		} else {
			_ = sub.Close()
		}
	})

	for {
		select {
		case <-termsig:
			return
		case <-time.After(timeout):
			t.Failed()
			return
		}
	}
}
