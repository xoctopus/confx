package envconf

import (
	"encoding"
	"fmt"
	"reflect"
)

var (
	TextMarshallerT   = reflect.TypeFor[encoding.TextMarshaler]()
	TextUnmarshallerT = reflect.TypeFor[encoding.TextUnmarshaler]()
)

type Code int8

const (
	E_DEC__INVALID_VALUE Code = iota + 1
	E_DEC__INVALID_VALUE_CANNOT_SET
	E_DEC__INVALID_MAP_KEY_TYPE
	E_DEC__FAILED_UNMARSHAL
	E_ENC__INVALID_MAP_KEY_TYPE
	E_ENC__INVALID_MAP_KEY_VALUE
	E_ENC__INVALID_ENV_KEY
	E_ENC__FAILED_MARSHAL
	E_ENC__DUPLICATE_GROUP_KEY
)

func (c Code) Message() string {
	prefix := fmt.Sprintf("[envconf.Error:%d] ", c)
	switch c {
	case E_DEC__INVALID_VALUE:
		return prefix + "got cannot set value."
	case E_DEC__INVALID_VALUE_CANNOT_SET:
		return prefix + "got nil value."
	case E_DEC__INVALID_MAP_KEY_TYPE:
		return prefix + "got invalid map key type, expect alphabet string or positive integer."
	case E_DEC__FAILED_UNMARSHAL:
		return prefix + "failed to unmarshal."
	case E_ENC__INVALID_MAP_KEY_TYPE:
		return prefix + "got invalid map key type, expect alphabet string or positive integer."
	case E_ENC__INVALID_MAP_KEY_VALUE:
		return prefix + "got invalid map key value, expect alphabet string or positive integer."
	case E_ENC__INVALID_ENV_KEY:
		return prefix + "got invalid env key, expect alphabet string."
	case E_ENC__FAILED_MARSHAL:
		return prefix + "failed to marshal."
	case E_ENC__DUPLICATE_GROUP_KEY:
		return prefix + "group key duplicated."
	default:
		return prefix + "unknown code."
	}
}
