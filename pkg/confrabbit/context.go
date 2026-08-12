package confrabbit

import "github.com/xoctopus/confx/pkg/types/mq"

var (
	With  = mq.With[ProducerMessage, ConsumerMessage]
	From  = mq.From[ProducerMessage, ConsumerMessage]
	Must  = mq.Must[ProducerMessage, ConsumerMessage]
	Carry = mq.Carry[ProducerMessage, ConsumerMessage]
)
