package confpulsar

// ecode presents
// +genx:code
type ecode int8

const (
	ECODE_UNDEFINED              ecode = iota
	ECODE__PARSE_MESSAGE               // failed to parse message
	ECODE__HANDLER_PANICKED            // subscriber handler panicked
	ECODE__SUBSCRIPTION_CANCELED       // subscription canceled
	ECODE__CLIENT_CLOSED               // client closed
	ECODE__SUBSCRIBER_CLOSED           // subscriber closed
	ECODE__PUBLISHER_CLOSED            // publisher closed
	ECODE__PUB_INVALID_MESSAGE         // publish invalid message
)
