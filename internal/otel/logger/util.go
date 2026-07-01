package logger

import (
	"fmt"
	"log/slog"
	"path"
	"runtime"
	"strconv"
	"time"

	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/slicex"
	"go.opentelemetry.io/otel/attribute"
	otelapilogger "go.opentelemetry.io/otel/log"

	"github.com/xoctopus/confx/internal/otel/consts"
	"github.com/xoctopus/confx/internal/otel/exporter"
)

func normalize(kvs []any) []otelapilogger.KeyValue {
	keyValues := make([]otelapilogger.KeyValue, 0, len(kvs))

	for i := 0; i < len(kvs); i++ {
		switch x := kvs[i].(type) {
		case []slog.Attr:
			keyValues = append(keyValues, slicex.Mapping(x, func(e slog.Attr) otelapilogger.KeyValue {
				return otelapilogger.KeyValue{
					Key:   e.Key,
					Value: exporter.LogAnyValue(e.Value.Any()),
				}
			})...)
		case slog.Attr:
			keyValues = append(keyValues, otelapilogger.KeyValue{
				Key:   x.Key,
				Value: exporter.LogAnyValue(x.Value.Any()),
			})
		case []attribute.KeyValue:
			keyValues = append(keyValues, slicex.M(x, func(e attribute.KeyValue) otelapilogger.KeyValue {
				return otelapilogger.KeyValue{
					Key:   string(e.Key),
					Value: exporter.LogAnyValue(e.Value.AsInterface()),
				}
			})...)
		case attribute.KeyValue:
			keyValues = append(keyValues, otelapilogger.KeyValue{
				Key:   string(x.Key),
				Value: exporter.LogAnyValue(x.Value.AsInterface()),
			})
		case string:
			// "key", value
			if i+1 < len(kvs) {
				i++
				keyValues = append(keyValues, otelapilogger.KeyValue{
					Key:   x,
					Value: exporter.LogAnyValue(kvs[i]),
				})
			}
		default:
			panic(fmt.Errorf("unsupported log attr values %T", x))
		}
	}

	return keyValues
}

func severity(lv logx.LogLevel) otelapilogger.Severity {
	switch lv {
	case logx.LogLevelDebug:
		return otelapilogger.SeverityDebug
	case logx.LogLevelInfo:
		return otelapilogger.SeverityInfo
	case logx.LogLevelWarn:
		return otelapilogger.SeverityWarn
	default:
		return otelapilogger.SeverityError
	}
}

func GetSource(skip int) Source {
	pc, _, _, _ := runtime.Caller(skip + 1)
	fs := runtime.CallersFrames([]uintptr{pc})
	f, _ := fs.Next()

	return Source{
		Function: f.Function,
		File:     f.File,
		Line:     f.Line,
	}
}

type Source slog.Source

func (s Source) AsKeyValues() []otelapilogger.KeyValue {
	return []otelapilogger.KeyValue{
		otelapilogger.String(consts.KEY_SOURCE_FUNC, s.Function),
		otelapilogger.String(consts.KEY_SOURCE_FILE, fmt.Sprintf("%s:%d", path.Base(s.File), s.Line)),
	}
}

func sprintf(format string, args ...any) fmt.Stringer {
	return &printer{format: format, args: args}
}

type printer struct {
	format string
	args   []any
}

func (p *printer) String() string {
	if len(p.args) == 0 {
		return p.format
	}
	return fmt.Sprintf(p.format, p.args...)
}

var (
	sec = []byte("s")
	ms  = []byte("ms")
	us  = []byte("μs")
	ns  = []byte("ns")
)

func cost(d time.Duration) string {
	var buf [16]byte
	b := buf[:0]

	if d >= time.Second {
		b = strconv.AppendFloat(b, d.Seconds(), 'f', 2, 64)
		b = append(b, sec...)
	} else if d >= time.Millisecond {
		val := float64(d) / float64(time.Millisecond)
		b = strconv.AppendFloat(b, val, 'f', 2, 64)
		b = append(b, ms...)
	} else if d >= time.Microsecond {
		val := float64(d) / float64(time.Microsecond)
		b = strconv.AppendFloat(b, val, 'f', 2, 64)
		b = append(b, us...)
	} else {
		b = strconv.AppendInt(b, d.Nanoseconds(), 10)
		b = append(b, ns...)
	}

	return string(b)
}
