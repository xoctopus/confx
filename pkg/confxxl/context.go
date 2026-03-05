package confxxl

import (
	"context"

	"cgtech.gitlab.com/saitox/x/contextx"
)

type Registry interface {
	Close() error

	RegisterJob(executor string, job string, fn JobHandler) error
	IsActive(executor string) bool
}

type (
	JobHandler  func(context.Context, *TriggerRequest) error
	JobCallback func(*TriggerRequest, error)
)

type tCtxRegistry struct{}

var (
	From  = contextx.From[tCtxRegistry, Registry]
	Must  = contextx.Must[tCtxRegistry, Registry]
	With  = contextx.With[tCtxRegistry, Registry]
	Carry = contextx.Carry[tCtxRegistry, Registry]
)
