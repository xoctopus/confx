package exporter

import (
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/log"
)

func LogValue(v log.Value) any {
	switch v.Kind() {
	case log.KindBool:
		return v.AsBool()
	case log.KindFloat64:
		return v.AsFloat64()
	case log.KindInt64:
		return v.AsInt64()
	case log.KindString:
		return v.AsString()
	case log.KindBytes:
		return v.AsBytes()
	case log.KindSlice:
		list := v.AsSlice()
		values := make([]any, len(list))
		for i := range list {
			values[i] = LogValue(list[i])
		}
		return values
	case log.KindMap:
		values := map[string]any{}
		for _, k := range v.AsMap() {
			values[k.Key] = LogValue(k.Value)
		}
		return values
	default:
		return nil
	}
}

func LogAnyValue(value any) log.Value {
	switch x := value.(type) {
	case time.Time:
		return log.StringValue(slog.TimeValue(x).String())
	case time.Duration:
		return log.StringValue(slog.DurationValue(x).String())
	case fmt.Stringer:
		return log.StringValue(x.String())
	case []byte:
		return log.BytesValue(x)
	case string:
		return log.StringValue(x)
	case uint:
		return log.Int64Value(int64(x))
	case uint8:
		return log.Int64Value(int64(x))
	case uint16:
		return log.Int64Value(int64(x))
	case uint32:
		return log.Int64Value(int64(x))
	case int:
		return log.Int64Value(int64(x))
	case int8:
		return log.Int64Value(int64(x))
	case int16:
		return log.Int64Value(int64(x))
	case int32:
		return log.Int64Value(int64(x))
	case int64:
		return log.Int64Value(x)
	case float32:
		return log.Float64Value(float64(x))
	case float64:
		return log.Float64Value(x)
	case bool:
		return log.BoolValue(x)
	case []any:
		values := make([]log.Value, len(x))
		for i, item := range x {
			values[i] = LogAnyValue(item)
		}
		return log.SliceValue(values...)
	case map[string]any:
		kvs := make([]log.KeyValue, 0, len(x))
		for k, v := range x {
			kvs = append(kvs, log.KeyValue{
				Key:   k,
				Value: LogAnyValue(v),
			})
		}
		return log.MapValue(kvs...)
	default:
		if u, ok := x.(interface{ Unwrap() any }); ok {
			return LogAnyValue(u.Unwrap())
		}
		return log.StringValue(slog.AnyValue(x).String())
	}
}
