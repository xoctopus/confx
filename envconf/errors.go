package envconf

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
)

var (
	TextMarshallerT   = reflect.TypeFor[encoding.TextMarshaler]()
	TextUnmarshallerT = reflect.TypeFor[encoding.TextUnmarshaler]()
)

type errCode int8

const (
	E_DEC__INVALID_VALUE errCode = iota + 1
	E_DEC__INVALID_VALUE_CANNOT_SET
	E_DEC__INVALID_MAP_KEY_TYPE
	E_DEC__FAILED_UNMARSHAL
	E_ENC__INVALID_MAP_KEY_TYPE
	E_ENC__INVALID_MAP_KEY_VALUE
	E_ENC__INVALID_ENV_KEY
	E_ENC__FAILED_MARSHAL
	E_ENC__DUPLICATE_GROUP_KEY
)

var errMessages = map[errCode]string{
	E_DEC__INVALID_VALUE:            "when decoding at `%s`, got cannot set value",
	E_DEC__INVALID_VALUE_CANNOT_SET: "when decoding at `%s`, got nil value",
	E_DEC__INVALID_MAP_KEY_TYPE:     "when decoding at `%s`, got invalid map key type, expect alphabet string or positive integer. but got:",
	E_DEC__FAILED_UNMARSHAL:         "when decoding at `%s`, failed to unmarshal",
	E_ENC__INVALID_MAP_KEY_TYPE:     "when encoding at `%s`, got invalid map key type, expect alphabet string or positive integer. but got:",
	E_ENC__INVALID_MAP_KEY_VALUE:    "when encoding at `%s`, got invalid map key value, expect alphabet string or positive integer. but got:",
	E_ENC__INVALID_ENV_KEY:          "when encoding at `%s`, got invalid env key, expect alphabet string. but got:",
	E_ENC__FAILED_MARSHAL:           "when encoding at `%s`, failed to marshal.",
	E_ENC__DUPLICATE_GROUP_KEY:      "when encoding at `%s`, group key duplicated:",
}

func NewError(pos *PathWalker, code errCode) error {
	return NewErrorf(pos, code, "")
}

func NewErrorW(pos *PathWalker, code errCode, wrap error) error {
	if wrap == nil {
		return nil
	}
	return NewErrorf(pos, code, "[%v]", wrap)
}

func NewErrorf(pos *PathWalker, code errCode, msg string, args ...any) error {
	position := ""
	if pos != nil {
		position = pos.String()
	}
	return &Error{
		pos:  position,
		code: code,
		msg:  fmt.Sprintf(msg, args...),
	}
}

type Error struct {
	pos  string
	code errCode
	msg  string
}

func (e *Error) Error() string {
	if len(e.msg) > 0 {
		return fmt.Sprintf(errMessages[e.code]+" "+e.msg, e.pos)
	}
	return fmt.Sprintf(errMessages[e.code], e.pos)
}

func (e *Error) Is(target error) bool {
	var err *Error
	ok := errors.As(target, &err)
	return ok &&
		e.code == err.code &&
		e.pos == err.pos
}
