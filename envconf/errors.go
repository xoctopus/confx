package envconf

import (
	"encoding"
	"errors"
	"reflect"
)

var (
	rtTextUnmarshaller = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	rtTextMarshaller   = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
)

func NewErrMarshalUnexpectMapKeyType(t reflect.Type) *ErrUnexpectMapKeyType {
	return &ErrUnexpectMapKeyType{
		rtpe: t,
		when: "marshal",
	}
}

func NewErrUnmarshalUnexpectMapKeyType(t reflect.Type) *ErrUnexpectMapKeyType {
	return &ErrUnexpectMapKeyType{
		rtpe: t,
		when: "unmarshal",
	}
}

type ErrUnexpectMapKeyType struct {
	rtpe reflect.Type
	when string
}

func (e *ErrUnexpectMapKeyType) Is(target error) bool {
	var x *ErrUnexpectMapKeyType
	if errors.As(target, &x) {
		return x.rtpe == e.rtpe && x.when == e.when
	}
	return false
}

func (e *ErrUnexpectMapKeyType) Error() string {
	msg := e.when + ": unexpected map key type, expect `integer` or `string`, but got "
	if e.rtpe == nil {
		return msg + "nil type"
	}
	return msg + e.rtpe.String()
}

type ErrUnexpectMapKeyValue string

func (e ErrUnexpectMapKeyValue) Error() string {
	return "unexpect map key, expect alphabet only, but got `" + string(e) + "`"
}

func NewInvalidValueErr() *ErrInvalidDecodeValue {
	return &ErrInvalidDecodeValue{rtpe: nil, reason: "invalid value"}
}

func NewCannotSetErr(t reflect.Type) *ErrInvalidDecodeValue {
	return &ErrInvalidDecodeValue{rtpe: t, reason: "cannot set"}
}

func NewInvalidDecodeError(t reflect.Type, reason string) *ErrInvalidDecodeValue {
	return &ErrInvalidDecodeValue{rtpe: t, reason: reason}
}

type ErrInvalidDecodeValue struct {
	rtpe   reflect.Type
	reason string
}

func (e *ErrInvalidDecodeValue) Is(target error) bool {
	var x *ErrInvalidDecodeValue
	if errors.As(target, &x) {
		return x.rtpe == e.rtpe && x.reason == e.reason
	}
	return false
}

func (e *ErrInvalidDecodeValue) Error() string {
	if e.rtpe == nil {
		return "decode(nil): " + e.reason
	}
	return "decode(`" + e.rtpe.String() + "`): " + e.reason
}
