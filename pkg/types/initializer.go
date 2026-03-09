package types

import (
	"context"
	"errors"
	"reflect"

	"github.com/xoctopus/x/reflectx"
)

type (
	_Initializer                   interface{ Init() }
	_InitializerWithError          interface{ Init() error }
	_InitializerByContext          interface{ Init(context.Context) }
	_InitializerByContextWithError interface{ Init(context.Context) error }
)

func CanBeInitialized(initializer any) bool {
	switch v := initializer.(type) {
	case _Initializer, _InitializerWithError, _InitializerByContext, _InitializerByContextWithError:
		return true
	case reflect.Value:
		v = reflectx.IndirectNew(initializer)
		if v == reflectx.InvalidValue {
			return false
		}
		if v.CanInterface() {
			if CanBeInitialized(v.Interface()) {
				return true
			}
		}
		if v.CanAddr() {
			if v.Addr().CanInterface() {
				if CanBeInitialized(v.Addr().Interface()) {
					return true
				}
			}
		}
		return false
	default:
		return false
	}
}

var ErrInvalidInitializerValue = errors.New("invalid initializer")

func InitByContext(ctx context.Context, initializer any) error {
	switch v := initializer.(type) {
	case _Initializer:
		v.Init()
		return nil
	case _InitializerWithError:
		return v.Init()
	case _InitializerByContext:
		v.Init(ctx)
		return nil
	case _InitializerByContextWithError:
		return v.Init(ctx)
	case reflect.Value:
		v = reflectx.IndirectNew(initializer)
		if v == reflectx.InvalidValue {
			return ErrInvalidInitializerValue
		}
		if v.CanInterface() {
			if CanBeInitialized(v.Interface()) {
				return InitByContext(ctx, v.Interface())
			}
		}
		if v.CanAddr() {
			if v.Addr().CanInterface() {
				if CanBeInitialized(v.Addr().Interface()) {
					return InitByContext(ctx, v.Addr().Interface())
				}
			}
		}
		return nil
	default:
		return nil
	}
}

func Init(initializer any) error {
	return InitByContext(context.Background(), initializer)
}
