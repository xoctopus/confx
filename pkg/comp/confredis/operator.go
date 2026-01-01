package confredis

import (
	"context"

	"github.com/gomodule/redigo/redis"
	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/misc/must"
)

type Operator interface {
	// Key returns key with prefix
	Key(key string) string
	// Get redis connection
	Get() redis.Conn

	Executor
}

type Executor interface {
	Exec(string, ...any) (any, error)
}

type MultiExecutor interface {
	// ExecCmd to execute multi redis command
	// ExecCmd(cmd *Cmd, others ...*Cmd) (any, error)
}

type tCtxOperator struct{}

func From(ctx context.Context) (Operator, bool) {
	o, ok := ctx.Value(tCtxOperator{}).(Operator)
	return o, ok
}

func MustFrom(ctx context.Context) Operator {
	return must.BeTrueV(From(ctx))
}

func With(ctx context.Context, o Operator) context.Context {
	return context.WithValue(ctx, tCtxOperator{}, o)
}

func Carrier(o Operator) contextx.Carrier {
	return func(ctx context.Context) context.Context {
		return With(ctx, o)
	}
}
