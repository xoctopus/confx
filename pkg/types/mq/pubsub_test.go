package mq_test

import (
	"context"
	"testing"

	. "cgtech.gitlab.com/saitox/x/testx"

	"cgtech.gitlab.com/saitox/confx/pkg/confpulsar"
	"cgtech.gitlab.com/saitox/confx/pkg/types/mq"
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
