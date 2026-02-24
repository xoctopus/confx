package mq_test

import (
	"context"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/confpulsar"
	"github.com/xoctopus/confx/pkg/types/mq"
)

func TestInjector(t *testing.T) {
	ep := &confpulsar.Endpoint{}
	ps := mq.Must[confpulsar.ProducerMessage, confpulsar.ConsumerMessage](
		mq.Carry(ep)(context.Background()),
	)
	Expect(t, ps != nil, BeTrue())

	ps, _ = mq.From[confpulsar.ProducerMessage, confpulsar.ConsumerMessage](context.Background())
	Expect(t, ps == nil, BeTrue())

	ctx := mq.With(context.Background(), ep)
	ps, _ = mq.From[confpulsar.ProducerMessage, confpulsar.ConsumerMessage](ctx)
	Expect(t, ps != nil, BeTrue())
}
