package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/xoctopus/logx"
	otelapilogger "go.opentelemetry.io/otel/log"
	otelapitracer "go.opentelemetry.io/otel/trace"

	"github.com/xoctopus/confx/internal/otel/consts"
	"github.com/xoctopus/confx/internal/otel/providers"
)

type loggerc struct {
	ctx       context.Context
	provider  otelapilogger.LoggerProvider
	parent    otelapitracer.SpanID
	logger    otelapilogger.Logger
	enabled   logx.LogLevel
	timestamp time.Time
}

func (lc loggerc) Start(ctx context.Context, name string, parent otelapitracer.SpanID) loggerc {
	lc.ctx = ctx
	lc.timestamp = time.Now()
	lc.parent = parent
	lp, ok := providers.LoggerProviderFrom(lc.ctx)
	if !ok {
		lp = lc.provider
	}
	lc.logger = lp.Logger(name)

	return lc
}

func (lc *loggerc) get() otelapilogger.Logger {
	if lc.logger == nil {
		lp, ok := providers.LoggerProviderFrom(lc.ctx)
		if !ok {
			lp = lc.provider
		}
		return lp.Logger("")
	}
	return lc.logger
}

func (lc *loggerc) emit(lv logx.LogLevel, msg fmt.Stringer, kvs []otelapilogger.KeyValue) {
	var rec otelapilogger.Record

	switch lv {
	case logx.LogLevelDebug:
		rec.SetSeverity(otelapilogger.SeverityDebug)
	case logx.LogLevelInfo:
		rec.SetSeverity(otelapilogger.SeverityInfo)
	case logx.LogLevelWarn:
		rec.AddAttributes(GetSource(3).AsKeyValues()...)
		rec.SetSeverity(otelapilogger.SeverityWarn)
	case logx.LogLevelError:
		rec.AddAttributes(GetSource(3).AsKeyValues()...)
		rec.SetSeverity(otelapilogger.SeverityError)
	}

	if !lc.timestamp.IsZero() {
		rec.AddAttributes(otelapilogger.String(consts.KEY_COST, cost(time.Since(lc.timestamp))))
	}

	if len(kvs) > 0 {
		rec.AddAttributes(kvs...)
	}

	if lc.parent.IsValid() {
		rec.AddAttributes(otelapilogger.String(consts.KEY_TRACE_PARENT_SPAN_ID, lc.parent.String()))
	}

	rec.SetTimestamp(time.Now())
	rec.SetBody(otelapilogger.StringValue(msg.String()))

	lc.get().Emit(lc.ctx, rec)
}

func (lc *loggerc) info(lv logx.LogLevel, msg fmt.Stringer, kvs []otelapilogger.KeyValue) {
	if lv > lc.enabled {
		lc.emit(lv, msg, kvs)
	}
}

func (lc *loggerc) error(lv logx.LogLevel, err error, kvs []otelapilogger.KeyValue, post func(err error)) {
	if lv > lc.enabled {
		if err == nil {
			return
		}
		// TODO error stack
		lc.emit(lv, sprintf("%s", err), kvs)
		post(err)
	}
}
