package types_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	_ "unsafe"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/types"
)

var (
	errCloseWithError          = errors.New("close with error")
	errCloseByContextWithError = errors.New("close by context with error")
)

type Closable struct{}

func (i *Closable) Close() {}

type ClosableV struct{}

func (i ClosableV) Close() {}

type ClosableWithError struct{}

func (i *ClosableWithError) Close() error {
	return errCloseWithError
}

type ClosableByContext struct{}

func (i *ClosableByContext) Close(_ context.Context) {}

type ClosableByContextWithError struct{}

func (i *ClosableByContextWithError) Close(_ context.Context) error {
	return errCloseByContextWithError
}

func TestClose(t *testing.T) {
	for i, v := range [...]struct {
		val any
		err error
	}{
		{&Closable{}, nil},                                                           // 0
		{&ClosableWithError{}, errCloseWithError},                                    // 1
		{&ClosableByContext{}, nil},                                                  // 2
		{&ClosableByContextWithError{}, errCloseByContextWithError},                  // 3
		{reflect.ValueOf(&ClosableByContextWithError{}), errCloseByContextWithError}, // 4
		{&struct{}{}, nil},                                                           // 5
		{reflect.ValueOf((*Initializer)(nil)), types.ErrInvalidClosableValue},        // 6
		{reflect.ValueOf(&struct{ Closable }{}), nil},                                // 7
		{reflect.ValueOf(&struct{ v Closable }{}), nil},                              // 8
		{reflect.ValueOf(&ClosableV{}), nil},                                         // 9
	} {
		_ = i
		if v.err == nil {
			Expect(t, types.Close(v.val), Succeed())
		} else {
			Expect(t, types.Close(v.val), Equal(v.err))
		}
	}
}

func TestIsClosable(t *testing.T) {
	for i, v := range [...]struct {
		v   any
		can bool
	}{
		{&Closable{}, true},
		{reflect.ValueOf(&Closable{}), true},
		{Closable{}, false},
		{reflect.ValueOf(Closable{}), false},
		{ClosableV{}, true},
		{reflect.ValueOf(ClosableV{}), true},
		{reflect.ValueOf(&struct{ _v Closable }{}).Elem().Field(0), false},
		{reflect.ValueOf(&struct{ V_ Closable }{}).Elem().Field(0), true},
		{reflect.ValueOf(&struct{ _v *Closable }{}).Elem().Field(0), false},
		{reflect.ValueOf(&struct{ V_ *Closable }{}).Elem().Field(0), true},
	} {
		_ = i
		can := types.IsClosable(v.v)
		Expect(t, v.can, Equal(can))
	}
}
