package exporter

import (
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

func LogValue(v attribute.Value) any {
	switch v.Type() {
	case attribute.BOOL:
		return v.AsBool()
	case attribute.FLOAT64:
		return v.AsFloat64()
	case attribute.INT64:
		return v.AsInt64()
	case attribute.STRING:
		return v.AsString()
	case attribute.BYTESLICE:
		return v.AsByteSlice()
	case attribute.SLICE:
		list := v.AsSlice()
		values := make([]any, len(list))
		for i := range list {
			values[i] = LogValue(list[i])
		}
		return values
	case attribute.MAP:
		values := map[string]any{}
		for _, k := range v.AsMap() {
			values[string(k.Key)] = LogValue(k.Value)
		}
		return values
	default:
		return nil
	}
}

func LogAnyValue(value any) attribute.Value {
	switch x := value.(type) {
	case time.Time:
		return attribute.StringValue(slog.TimeValue(x).String())
	case time.Duration:
		return attribute.StringValue(slog.DurationValue(x).String())
	case fmt.Stringer:
		return attribute.StringValue(x.String())
	case []byte:
		return attribute.ByteSliceValue(x)
	case string:
		return attribute.StringValue(x)
	case uint:
		return attribute.Int64Value(int64(x))
	case uint8:
		return attribute.Int64Value(int64(x))
	case uint16:
		return attribute.Int64Value(int64(x))
	case uint32:
		return attribute.Int64Value(int64(x))
	case int:
		return attribute.Int64Value(int64(x))
	case int8:
		return attribute.Int64Value(int64(x))
	case int16:
		return attribute.Int64Value(int64(x))
	case int32:
		return attribute.Int64Value(int64(x))
	case int64:
		return attribute.Int64Value(x)
	case float32:
		return attribute.Float64Value(float64(x))
	case float64:
		return attribute.Float64Value(x)
	case bool:
		return attribute.BoolValue(x)
	case []any:
		values := make([]attribute.Value, len(x))
		for i, item := range x {
			values[i] = LogAnyValue(item)
		}
		return attribute.SliceValue(values...)
	case map[string]any:
		kvs := make([]attribute.KeyValue, 0, len(x))
		for k, v := range x {
			kvs = append(kvs, attribute.KeyValue{
				Key:   attribute.Key(k),
				Value: LogAnyValue(v),
			})
		}
		return attribute.MapValue(kvs...)
	default:
		if u, ok := x.(interface{ Unwrap() any }); ok {
			return LogAnyValue(u.Unwrap())
		}
		return attribute.StringValue(slog.AnyValue(x).String())
	}
}
