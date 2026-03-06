package kv

import (
	"context"

	"github.com/xoctopus/x/contextx"
)

type Executor interface {
	// Key returns cache key with given key.
	Key(string) string
	// Exec executes cache command
	Exec(context.Context, string, ...any) (any, error)
}

type tCtxExecutor struct{}

var (
	With  = contextx.With[tCtxExecutor, Executor]
	From  = contextx.From[tCtxExecutor, Executor]
	Must  = contextx.Must[tCtxExecutor, Executor]
	Carry = contextx.Carry[tCtxExecutor, Executor]
)
