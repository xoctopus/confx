package types

import (
	"context"
	"errors"
	"reflect"

	"github.com/xoctopus/x/reflectx"
)

type (
	Closable interface {
		Close()
	}

	ClosableWithError interface {
		Close() error
	}

	ClosableByContext interface {
		Close(context.Context)
	}

	ClosableByContextWithError interface {
		Close(context.Context) error
	}

	Shutdownable interface {
		Shutdown(ctx context.Context) error
	}
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
	case Closable, ClosableWithError, ClosableByContext, ClosableByContextWithError, Shutdownable:
		return true
	default:
		return false
	}
}

func Close(closable any) error {
	return CloseByContext(context.Background(), closable)
}

var ErrSkipClosing = errors.New("skip closing")

func CloseByContext(ctx context.Context, closable any) error {
	switch x := closable.(type) {
	case reflect.Value:
		x = reflectx.IndirectNew(closable)
		if x == reflectx.InvalidValue {
			return ErrSkipClosing
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
	case Closable:
		x.Close()
		return nil
	case ClosableWithError:
		return x.Close()
	case ClosableByContext:
		x.Close(ctx)
		return nil
	case ClosableByContextWithError:
		return x.Close(ctx)
	case Shutdownable:
		return x.Shutdown(ctx)
	default:
		return nil
	}
}
