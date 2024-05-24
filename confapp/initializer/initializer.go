package initializer

import (
	"context"
	"errors"
	"reflect"

	"github.com/sincospro/x/reflectx"
)

type (
	_Initializer        interface{ Init() }
	_WithError          interface{ Init() error }
	_ByContext          interface{ Init(context.Context) }
	_ByContextWithError interface{ Init(context.Context) error }
)

func CanBeInitialized(initializer any) bool {
	switch v := initializer.(type) {
	case _Initializer, _WithError, _ByContext, _ByContextWithError:
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

var ErrInvalidValue = errors.New("invalid value")

func InitByContext(ctx context.Context, initializer any) error {
	switch v := initializer.(type) {
	case _Initializer:
		v.Init()
		return nil
	case _WithError:
		return v.Init()
	case _ByContext:
		v.Init(ctx)
		return nil
	case _ByContextWithError:
		return v.Init(ctx)
	case reflect.Value:
		v = reflectx.IndirectNew(initializer)
		if v == reflectx.InvalidValue {
			return ErrInvalidValue
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
