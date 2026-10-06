package confrabbit_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/wagslane/go-rabbitmq"
	"github.com/xoctopus/x/codex"
	. "github.com/xoctopus/x/testx"
	"github.com/xoctopus/x/testx/bdd"

	"github.com/xoctopus/confx/hack"
	. "github.com/xoctopus/confx/pkg/confrabbit"
	"github.com/xoctopus/confx/pkg/types/mq"
)

const (
	producerTestDSN      = "amqp://guest:guest@localhost:5672"
	producerTestExchange = "confrabbit_test_producer"
)

func TopicFor(t testing.TB) string {
	return strings.ReplaceAll(t.Name(), "/", "_")
}

// Sink records payloads delivered to a queue bound to producerTestExchange.
type Sink struct {
	mu   sync.Mutex
	seen map[string]struct{}
}

func (s *Sink) add(payload []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seen[string(payload)] = struct{}{}
}

func (s *Sink) has(payloads ...string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range payloads {
		if _, ok := s.seen[p]; !ok {
			return false
		}
	}
	return true
}

func eventually(timeout time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return cond()
}

func NewTestProducer(t testing.TB, ctx context.Context, topic string, appliers ...mq.OptionApplier) mq.Producer[ProducerMessage] {
	t.Helper()
	// declared with the same flags as Subscribe, publishing to an undeclared
	// exchange gets the channel closed by broker and stalls until reconnected
	pub, err := Must(ctx).NewProducer(ctx, append([]mq.OptionApplier{
		WithPubTopic(topic),
		WithPubExchange(producerTestExchange, "direct"),
		WithRabbitPublisherOptions(
			rabbitmq.WithPublisherOptionsExchangeDeclare,
			rabbitmq.WithPublisherOptionsExchangeAutoDelete,
		),
	}, appliers...)...)
	Expect(t, err, Succeed())
	t.Cleanup(func() { _ = pub.Close() })
	return pub
}

// Subscribe binds an exclusive queue to producerTestExchange with topic as
// routing key, and returns once the binding is proven to deliver messages.
func Subscribe(t testing.TB, ctx context.Context, topic string) *Sink {
	t.Helper()
	s := &Sink{seen: map[string]struct{}{}}
	sub, err := Must(ctx).NewConsumer(ctx,
		WithSubQueue(ulid.Make().String()),
		WithSubExchange(producerTestExchange, "direct"),
		WithSubRoutingKey(topic),
		WithSubWorker(1),
		WithRabbitConsumerOptions(
			rabbitmq.WithConsumerOptionsExchangeDeclare,
			rabbitmq.WithConsumerOptionsExchangeAutoDelete,
			rabbitmq.WithConsumerOptionsQueueAutoDelete,
			rabbitmq.WithConsumerOptionsQueueExclusive,
		),
	)
	Expect(t, err, Succeed())
	t.Cleanup(func() { _ = sub.Close() })
	go func() {
		_ = sub.Run(ctx, func(_ context.Context, m ConsumerMessage) error {
			s.add(m.Payload())
			return nil
		})
	}()

	// the queue and binding are declared asynchronously by Run
	probe := NewTestProducer(t, ctx, topic, WithSyncPublish())
	payload := "probe_" + ulid.Make().String()
	ready := eventually(5*time.Second, func() bool {
		_, _ = probe.Publish(ctx, topic, []byte(payload))
		return s.has(payload)
	})
	Expect(t, ready, BeTrue())
	return s
}

// PublishAsync publishes n messages through an async producer and waits for
// all callbacks. If cancelCaller, each caller context is canceled right after
// Publish returns, simulating a request scoped context.
func PublishAsync(
	t testing.TB, ctx context.Context, topic string, n int, cancelCaller bool,
	appliers ...mq.OptionApplier,
) (payloads []string, errs []error) {
	t.Helper()
	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)
	pub := NewTestProducer(t, ctx, topic, append(appliers,
		WithPublishCallback(func(_ ProducerMessage, err error) {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
			wg.Done()
		}),
	)...)

	wg.Add(n)
	for range n {
		payload := ulid.Make().String()
		payloads = append(payloads, payload)
		caller, cancel := context.WithCancel(ctx)
		_, err := pub.Publish(caller, topic, []byte(payload))
		Expect(t, err, Succeed())
		if cancelCaller {
			cancel()
		} else {
			defer cancel()
		}
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("async publish callbacks not completed in time")
	}
	return payloads, errs
}

func count(errs []error, pred func(error) bool) (n int) {
	for _, err := range errs {
		if pred(err) {
			n++
		}
	}
	return n
}

func isNil(err error) bool { return err == nil }

func isTimeout(err error) bool { return codex.IsCode(err, ECODE__PUB_TIMEOUT) }

func TestProducer_PublishMessage(t *testing.T) {
	bdd.From(t).Given("sync producer", func(t bdd.T) {
		var (
			ctx   = hack.WithRabbitMQ(hack.Context(t), t, producerTestDSN)
			topic = TopicFor(t)
			recv  = Subscribe(t, ctx, topic)
		)

		t.When("publish without timeout", func(t bdd.T) {
			pub := NewTestProducer(t, ctx, topic, WithSyncPublish(), WithPubTimeout(0))
			payload := ulid.Make().String()
			_, err := pub.Publish(ctx, topic, []byte(payload))
			t.Then("succeeded and delivered",
				bdd.Succeed(err),
				bdd.BeTrue(eventually(5*time.Second, func() bool { return recv.has(payload) })),
			)
		})

		t.When("publish with timeout", func(t bdd.T) {
			pub := NewTestProducer(t, ctx, topic, WithSyncPublish(), WithPubTimeout(2*time.Second))
			payload := ulid.Make().String()
			_, err := pub.Publish(ctx, topic, []byte(payload))
			t.Then("succeeded and delivered",
				bdd.Succeed(err),
				bdd.BeTrue(eventually(5*time.Second, func() bool { return recv.has(payload) })),
			)
		})

		t.When("publish with timeout in confirm mode", func(t bdd.T) {
			pub := NewTestProducer(t, ctx, topic, WithSyncPublish(), WithPubConfirm(), WithPubTimeout(2*time.Second))
			payload := ulid.Make().String()
			_, err := pub.Publish(ctx, topic, []byte(payload))
			t.Then("succeeded after broker ack and delivered",
				bdd.Succeed(err),
				bdd.BeTrue(eventually(5*time.Second, func() bool { return recv.has(payload) })),
			)
		})

		t.When("publish exceeds timeout", func(t bdd.T) {
			pub := NewTestProducer(t, ctx, topic, WithSyncPublish(), WithPubTimeout(time.Nanosecond))
			_, err := pub.Publish(ctx, topic, []byte(ulid.Make().String()))
			t.Then("failed with publish timeout", bdd.IsCodeError(err, ECODE__PUB_TIMEOUT))
		})

		t.When("publish exceeds timeout in confirm mode", func(t bdd.T) {
			pub := NewTestProducer(t, ctx, topic, WithSyncPublish(), WithPubConfirm(), WithPubTimeout(time.Nanosecond))
			_, err := pub.Publish(ctx, topic, []byte(ulid.Make().String()))
			t.Then("failed with publish timeout", bdd.IsCodeError(err, ECODE__PUB_TIMEOUT))
		})

		t.When("caller context already canceled", func(t bdd.T) {
			pub := NewTestProducer(t, ctx, topic, WithSyncPublish(), WithPubTimeout(2*time.Second))
			caller, cancel := context.WithCancel(ctx)
			cancel()
			_, err := pub.Publish(caller, topic, []byte(ulid.Make().String()))
			t.Then("failed with caller cancellation instead of timeout",
				bdd.IsError(context.Canceled, err),
				bdd.BeFalse(isTimeout(err)),
			)
		})
	})

	bdd.From(t).Given("async producer", func(t bdd.T) {
		const n = 100
		var (
			ctx   = hack.WithRabbitMQ(hack.Context(t), t, producerTestDSN)
			topic = TopicFor(t)
			recv  = Subscribe(t, ctx, topic)
		)

		t.When("publish without timeout and caller context canceled after return", func(t bdd.T) {
			payloads, errs := PublishAsync(t, ctx, topic, n, true, WithPubTimeout(0))
			t.Then("all callbacks succeeded and delivered",
				bdd.HaveLen(errs, n),
				bdd.Equal(count(errs, isNil), n),
				bdd.BeTrue(eventually(5*time.Second, func() bool { return recv.has(payloads...) })),
			)
		})

		t.When("publish with timeout and caller context canceled after return", func(t bdd.T) {
			payloads, errs := PublishAsync(t, ctx, topic, n, true, WithPubTimeout(2*time.Second))
			t.Then("all callbacks succeeded and delivered",
				bdd.HaveLen(errs, n),
				bdd.Equal(count(errs, isNil), n),
				bdd.BeTrue(eventually(5*time.Second, func() bool { return recv.has(payloads...) })),
			)
		})

		t.When("publish with timeout in confirm mode", func(t bdd.T) {
			payloads, errs := PublishAsync(t, ctx, topic, n, true, WithPubConfirm(), WithPubTimeout(2*time.Second))
			t.Then("all callbacks succeeded after broker ack and delivered",
				bdd.HaveLen(errs, n),
				bdd.Equal(count(errs, isNil), n),
				bdd.BeTrue(eventually(5*time.Second, func() bool { return recv.has(payloads...) })),
			)
		})

		t.When("publish exceeds timeout", func(t bdd.T) {
			_, errs := PublishAsync(t, ctx, topic, n, false, WithPubTimeout(time.Nanosecond))
			t.Then("all callbacks got publish timeout",
				bdd.HaveLen(errs, n),
				bdd.Equal(count(errs, isTimeout), n),
			)
		})

		t.When("publish exceeds timeout in confirm mode", func(t bdd.T) {
			_, errs := PublishAsync(t, ctx, topic, n, false, WithPubConfirm(), WithPubTimeout(time.Nanosecond))
			t.Then("all callbacks got publish timeout",
				bdd.HaveLen(errs, n),
				bdd.Equal(count(errs, isTimeout), n),
			)
		})
	})
}
