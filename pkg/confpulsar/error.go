package confpulsar

// Error presents error codes for confpulsar
// +genx:code
type Error int8

const (
	ERROR_UNDEFINED                Error = iota
	ERROR__CLI_CLOSED                    // client closed
	ERROR__CLI_INIT_ERROR                // client init failed
	ERROR__SUB_CLOSED                    // subscriber closed
	ERROR__SUB_BOOTED                    // subscriber is already booted
	ERROR__SUB_HANDLER_PANICKED          // subscriber handler panicked
	ERROR__SUB_PARSE_MESSAGE_ERROR       // subscriber failed to parse message
	ERROR__SUB_UNSUBSCRIBED              // subscriber unsubscribed
	ERROR__PUB_CLOSED                    // publisher closed
	ERROR__PUB_INVALID_MESSAGE           // publisher got invalid message
)
