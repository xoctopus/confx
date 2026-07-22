package confrabbit

// Error presents
// +genx:code
type Error int8

const (
	ECODE_UNDEFINED             Error = iota
	ECODE__CLI_CLOSED                 // client closed
	ECODE__SUB_CLOSED                 // subscriber closed
	ECODE__SUB_BOOTED                 // subscriber is already booted
	ECODE__SUB_HANDLER_PANICKED       // subscriber handler panicked
	ECODE__PUB_CLOSED                 // publisher closed
	ECODE__PUB_INVALID_MESSAGE        // publisher got invalid message
	ECODE__SUB_UNSUBSCRIBED           // subscriber unsubscribed
)
