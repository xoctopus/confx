package confrabbit

import (
	"context"
	"errors"

	"github.com/xoctopus/x/codex"
)

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
	ECODE__PUB_TIMEOUT                // publisher timed out before the broker accepted the message
	ECODE__PUB_NACKED                 // publisher got negative acknowledgement from the broker
)

func WrapPublishTimeoutError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return codex.Wrap(ECODE__PUB_TIMEOUT, err)
	}
	return err
}
