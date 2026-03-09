package types

import (
	"context"
	"errors"
	"reflect"

	"github.com/xoctopus/x/reflectx"
)

type (
	_Closable                   interface{ Close() }
	_ClosableWithError          interface{ Close() error }
	_ClosableByContext          interface{ Close(context.Context) }
	_ClosableByContextWithError interface{ Close(context.Context) error }
)

func IsClosable(closable any) bool {
	switch x := closable.(type) {
	case reflect.Value:
		x = reflectx.IndirectNew(closable)
		if x == reflectx.InvalidValue {
			return false
		}
		if x.CanInterface() {
			if IsClosable(x.Interface()) {
				return true
			}
		}
		if x.CanAddr() {
			if x.Addr().CanInterface() {
				if IsClosable(x.Addr().Interface()) {
					return true
				}
			}
		}
		return false
	case _Closable, _ClosableWithError, _ClosableByContext, _ClosableByContextWithError:
		return true
	default:
		return false
	}
}

func Close(closable any) error {
	return CloseByContext(context.Background(), closable)
}

var ErrInvalidClosableValue = errors.New("invalid closable")

func CloseByContext(ctx context.Context, closable any) error {
	switch x := closable.(type) {
	case reflect.Value:
		x = reflectx.IndirectNew(closable)
		if x == reflectx.InvalidValue {
			return ErrInvalidClosableValue
		}
		if x.CanInterface() {
			if IsClosable(x.Interface()) {
				return CloseByContext(ctx, x.Interface())
			}
		}
		if x.CanAddr() {
			if x.Addr().CanInterface() {
				if IsClosable(x.Addr().Interface()) {
					return CloseByContext(ctx, x.Addr().Interface())
				}
			}
		}
		return nil
	case _Closable:
		x.Close()
		return nil
	case _ClosableWithError:
		return x.Close()
	case _ClosableByContext:
		x.Close(ctx)
		return nil
	case _ClosableByContextWithError:
		return x.Close(ctx)
	default:
		return nil
	}
}
