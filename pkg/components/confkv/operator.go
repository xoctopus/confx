package confkv

import (
	"context"

	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/misc/must"
)

type Executor interface {
	Key(string) string
	Exec(string, ...any) (any, error)
}

type tCtxExecutor struct{}

func From(ctx context.Context) (Executor, bool) {
	o, ok := ctx.Value(tCtxExecutor{}).(Executor)
	return o, ok
}

func Must(ctx context.Context) Executor {
	return must.BeTrueV(From(ctx))
}

func With(ctx context.Context, o Executor) context.Context {
	return context.WithValue(ctx, tCtxExecutor{}, o)
}

func Carry(o Executor) contextx.Carrier {
	return func(ctx context.Context) context.Context {
		return With(ctx, o)
	}
}
