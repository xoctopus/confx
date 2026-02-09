package mq

import (
	"encoding"
	"encoding/json"
)

func MarshalV(v any) ([]byte, error) {
	switch x := v.(type) {
	case nil:
		return nil, nil
	case []byte:
		return x, nil
	case string:
		return []byte(x), nil
	case encoding.TextMarshaler:
		return x.MarshalText()
	case json.Marshaler:
		return x.MarshalJSON()
	default:
		return json.Marshal(x)
	}
}

func UnmarshalV(data []byte, v any) error {
	if v == nil && len(data) == 0 {
		return nil
	}
	switch x := v.(type) {
	case *[]byte:
		*x = data
		return nil
	case *string:
		*x = string(data)
		return nil
	case encoding.TextUnmarshaler:
		return x.UnmarshalText(data)
	case json.Unmarshaler:
		return x.UnmarshalJSON(data)
	default:
		return json.Unmarshal(data, x)
	}
}
