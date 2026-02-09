package kv

import (
	"github.com/xoctopus/x/contextx"
)

type Executor interface {
	Key(string) string
	Exec(string, ...any) (any, error)
}

type tCtxExecutor struct{}

var (
	With  = contextx.With[tCtxExecutor, Executor]
	From  = contextx.From[tCtxExecutor, Executor]
	Must  = contextx.Must[tCtxExecutor, Executor]
	Carry = contextx.Carry[tCtxExecutor, Executor]
)
