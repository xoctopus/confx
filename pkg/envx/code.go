package envx

import (
	"encoding"
	"reflect"
)

var (
	TextMarshallerT   = reflect.TypeFor[encoding.TextMarshaler]()
	TextUnmarshallerT = reflect.TypeFor[encoding.TextUnmarshaler]()
)

// Code defines env parsing error code
// +genx:code
// @def envx.Error
type Code int8

const (
	CODE_UNDEFINED                     Code = iota + 1
	CODE__DEC_INVALID_VALUE                 // cannot set value
	CODE__DEC_INVALID_VALUE_CANNOT_SET      // nil value
	CODE__DEC_INVALID_MAP_KEY_TYPE          // invalid map key type, expect alphabet string or positive integer
	CODE__DEC_FAILED_UNMARSHAL              // failed to unmarshal
	CODE__ENC_INVALID_MAP_KEY_TYPE          // invalid map key type, expect alphabet string or positive integer
	CODE__ENC_INVALID_MAP_KEY_VALUE         // invalid map key value, expect alphabet string or positive integer
	CODE__ENC_INVALID_ENV_KEY               // invalid env key, expect alphabet string
	CODE__ENC_FAILED_MARSHAL                // failed to marshal
	CODE__ENC_DUPLICATE_GROUP_KEY           // group key duplicated
)
