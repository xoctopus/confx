package types

import (
	"context"
	"reflect"

	"github.com/xoctopus/x/reflectx"
)

type Injectable interface {
	WithContext(ctx context.Context) context.Context
}

func IsInjectable(v any) bool {
	switch x := v.(type) {
	case Injectable:
		return true
	case reflect.Value:
		x = reflectx.IndirectNew(v)
		if x == reflectx.InvalidValue {
			return false
		}
		if x.CanInterface() {
			if IsInjectable(x.Interface()) {
				return true
			}
		}
		if x.CanAddr() {
			if x.Addr().CanInterface() {
				if IsInjectable(x.Addr().Interface()) {
					return true
				}
			}
		}
		return false
	default:
		return false
	}
}

func Inject(ctx context.Context, v any) context.Context {
	switch x := v.(type) {
	case Injectable:
		return x.WithContext(ctx)
	case reflect.Value:
		x = reflectx.IndirectNew(v)
		if x == reflectx.InvalidValue {
			return ctx
		}
		if x.CanInterface() {
			if IsInjectable(x.Interface()) {
				return Inject(ctx, x.Interface())
			}
		}
		if x.CanAddr() {
			if x.Addr().CanInterface() {
				if IsInjectable(x.Addr().Interface()) {
					return Inject(ctx, x.Addr().Interface())
				}
			}
		}
		return ctx
	default:
		return ctx
	}
}
