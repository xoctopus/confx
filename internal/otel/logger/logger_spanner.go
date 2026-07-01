package logger

import (
	"context"
	"time"

	otelapitracer "go.opentelemetry.io/otel/trace"

	"github.com/xoctopus/confx/internal/otel/providers"
)

type spanner struct {
	tp   otelapitracer.TracerProvider
	name string
	span otelapitracer.Span
}

func (c spanner) Start(ctx context.Context, name string) (spanner, context.Context) {
	tp, ok := providers.TracerProviderFrom(ctx)
	if !ok {
		tp = c.tp
	}
	cctx, span := tp.Tracer("").Start(ctx, c.name, otelapitracer.WithTimestamp(time.Now()))
	c.span = span
	c.name = name
	return c, cctx
}
