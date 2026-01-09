package confpulsar

// ecode presents
// +genx:code
type ecode int8

const (
	ECODE_UNDEFINED         ecode = iota
	ECODE__PARSE_MESSAGE          // failed to parse message
	ECODE__HANDLER_PANICKED       // subscriber handler panicked
)
