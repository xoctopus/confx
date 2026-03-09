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
	errInitializerWithError          = errors.New("initialize with error")
	errInitializerByContextWithError = errors.New("initialize by context with error")
)

type Initializer struct{}

func (i *Initializer) Init() {}

type InitializerV struct{}

func (i InitializerV) Init() {}

type InitializerWithError struct{}

func (i *InitializerWithError) Init() error {
	return errInitializerWithError
}

type InitializerByContext struct{}

func (i *InitializerByContext) Init(_ context.Context) {}

type InitializerByContextWithError struct{}

func (i *InitializerByContextWithError) Init(_ context.Context) error {
	return errInitializerByContextWithError
}

func TestInit(t *testing.T) {
	for i, v := range [...]struct {
		val any
		err error
	}{
		{&Initializer{}, nil},                                                                 // 0
		{&InitializerWithError{}, errInitializerWithError},                                    // 1
		{&InitializerByContext{}, nil},                                                        // 2
		{&InitializerByContextWithError{}, errInitializerByContextWithError},                  // 3
		{reflect.ValueOf(&InitializerByContextWithError{}), errInitializerByContextWithError}, // 4
		{&struct{}{}, nil},                                                                    // 5
		{reflect.ValueOf((*Initializer)(nil)), types.ErrInvalidInitializerValue},              // 6
		{reflect.ValueOf(&struct{ Initializer }{}), nil},                                      // 7
		{reflect.ValueOf(&struct{ v Initializer }{}), nil},                                    // 8
		{reflect.ValueOf(&InitializerV{}), nil},                                               // 9
	} {
		_ = i
		if v.err == nil {
			Expect(t, types.Init(v.val), Succeed())
		} else {
			Expect(t, types.Init(v.val), Equal(v.err))
		}
	}
}

func TestCanBeInitialized(t *testing.T) {
	for i, v := range [...]struct {
		v   any
		can bool
	}{
		{&Initializer{}, true},
		{reflect.ValueOf(&Initializer{}), true},
		{Initializer{}, false},
		{reflect.ValueOf(Initializer{}), false},
		{InitializerV{}, true},
		{reflect.ValueOf(InitializerV{}), true},
		{reflect.ValueOf(&struct{ _v Initializer }{}).Elem().Field(0), false},
		{reflect.ValueOf(&struct{ V_ Initializer }{}).Elem().Field(0), true},
		{reflect.ValueOf(&struct{ _v *Initializer }{}).Elem().Field(0), false},
		{reflect.ValueOf(&struct{ V_ *Initializer }{}).Elem().Field(0), true},
	} {
		_ = i
		can := types.CanBeInitialized(v.v)
		Expect(t, v.can, Equal(can))
	}
}
