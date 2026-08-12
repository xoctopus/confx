package types_test

import (
	"context"
	"reflect"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/types"
)

type injectableKey struct{}

type Injectable struct{}

func (i *Injectable) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, injectableKey{}, true)
}

type InjectableV struct{}

func (i InjectableV) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, injectableKey{}, true)
}

func injected(ctx context.Context) bool {
	v, _ := ctx.Value(injectableKey{}).(bool)
	return v
}

func TestInject(t *testing.T) {
	ctx := context.Background()

	for i, v := range [...]struct {
		val      any
		injected bool
	}{
		{&Injectable{}, true},                               // 0
		{reflect.ValueOf(&Injectable{}), true},              // 1
		{&struct{}{}, false},                                // 2
		{reflect.ValueOf((*Injectable)(nil)), false},        // 3
		{reflect.ValueOf(&struct{ Injectable }{}), true},    // 4
		{reflect.ValueOf(&struct{ v Injectable }{}), false}, // 5
		{reflect.ValueOf(&InjectableV{}), true},             // 6
		{InjectableV{}, true},                               // 7
		{reflect.ValueOf(InjectableV{}), true},              // 8
		{Injectable{}, false},                               // 9
		{reflect.ValueOf(Injectable{}), false},              // 10
	} {
		_ = i
		got := types.Inject(ctx, v.val)
		Expect(t, injected(got), Equal(v.injected))
	}
}

func TestIsInjectable(t *testing.T) {
	for i, v := range [...]struct {
		v   any
		can bool
	}{
		{&Injectable{}, true},
		{reflect.ValueOf(&Injectable{}), true},
		{Injectable{}, false},
		{reflect.ValueOf(Injectable{}), false},
		{InjectableV{}, true},
		{reflect.ValueOf(InjectableV{}), true},
		{reflect.ValueOf(&struct{ _v Injectable }{}).Elem().Field(0), false},
		{reflect.ValueOf(&struct{ V_ Injectable }{}).Elem().Field(0), true},
		{reflect.ValueOf(&struct{ _v *Injectable }{}).Elem().Field(0), false},
		{reflect.ValueOf(&struct{ V_ *Injectable }{}).Elem().Field(0), true},
	} {
		_ = i
		can := types.IsInjectable(v.v)
		Expect(t, v.can, Equal(can))
	}
}
