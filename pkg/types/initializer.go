package types

import (
	"context"
	"errors"
	"reflect"

	"github.com/xoctopus/x/reflectx"
)

type (
	Initializer interface {
		Init()
	}

	InitializerWithError interface {
		Init() error
	}

	InitializerByContext interface {
		Init(context.Context)
	}

	InitializerByContextWithError interface {
		Init(context.Context) error
	}
)

func CanBeInitialized(initializer any) bool {
	switch v := initializer.(type) {
	case Initializer, InitializerWithError, InitializerByContext, InitializerByContextWithError:
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

var ErrSkipInitializing = errors.New("skip initializing")

func InitByContext(ctx context.Context, initializer any) error {
	switch v := initializer.(type) {
	case Initializer:
		v.Init()
		return nil
	case InitializerWithError:
		return v.Init()
	case InitializerByContext:
		v.Init(ctx)
		return nil
	case InitializerByContextWithError:
		return v.Init(ctx)
	case reflect.Value:
		v = reflectx.IndirectNew(initializer)
		if v == reflectx.InvalidValue {
			return ErrSkipInitializing
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
