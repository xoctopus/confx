package logger

import (
	"context"
	"time"

	"github.com/xoctopus/logx"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"

	"github.com/xoctopus/confx/internal/otel/providers"
)

func NewLogger(ctx context.Context, lv logx.LogLevel) logx.Logger {
	tp, _ := providers.TracerProviderFrom(ctx)
	lp, _ := providers.LoggerProviderFrom(ctx)
	return &logger{
		spanner: spanner{tp: tp},
		loggerc: loggerc{
			ctx:      ctx,
			provider: lp,
			enabled:  lv,
		},
	}
}

type logger struct {
	spanner
	loggerc

	kvs []log.KeyValue
}

func (l *logger) With(kvs ...any) logx.Logger {
	if len(kvs) == 0 {
		return l
	}

	return &logger{
		spanner: l.spanner,
		loggerc: l.loggerc,
		kvs:     append(l.kvs, normalize(kvs)...),
	}
}

func (l *logger) Start(ctx context.Context, name string, kvs ...any) (context.Context, logx.Logger) {
	var parentID trace.SpanID

	parentSpan := trace.SpanContextFromContext(ctx)
	if parentSpan.HasSpanID() {
		parentID = parentSpan.SpanID()
	}

	spanCtx, c := l.spanner.Start(ctx, name)

	lgr := &logger{
		kvs:     append(l.kvs, normalize(kvs)...),
		spanner: spanCtx,
		loggerc: l.loggerc.Start(c, name, parentID),
	}

	return logx.With(c, lgr), lgr
}

func (l *logger) End() {
	now := time.Now()
	l.span(func(s trace.Span) {
		s.End(trace.WithTimestamp(now))
	})
}

func (l *logger) span(do func(s trace.Span)) {
	if span := l.spanner.span; span != nil {
		do(span)
	}
}

func (l *logger) Debug(f string, args ...any) {
	l.info(logx.LogLevelDebug, sprintf(f, args...), l.kvs)
}

func (l *logger) Info(f string, args ...any) {
	l.info(logx.LogLevelInfo, sprintf(f, args...), l.kvs)
}

func (l *logger) Warn(err error) {
	l.error(
		logx.LogLevelWarn,
		err, l.kvs,
		func(err error) {
			errMsg := err.Error()
			l.span(func(s trace.Span) {
				s.RecordError(err)
				s.SetStatus(codes.Error, errMsg)
			})
		},
	)
}

func (l *logger) Error(err error) {
	l.error(
		logx.LogLevelError,
		err, l.kvs,
		func(err error) {
			errMsg := err.Error()
			l.span(func(s trace.Span) {
				s.RecordError(err)
				s.SetStatus(codes.Error, errMsg)
			})
		},
	)
}
