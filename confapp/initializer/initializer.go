package initializer

import (
	"context"
	"reflect"
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
		return CanBeInitialized(v.Interface())
	default:
		return false
	}
}

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
		return InitByContext(ctx, v.Interface())
	default:
		return nil
	}
}

func Init(initializer any) error {
	return InitByContext(context.Background(), initializer)
}
