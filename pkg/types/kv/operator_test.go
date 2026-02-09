package kv_test

import (
	"context"
	"testing"

	"github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/confredis/v1"
	"github.com/xoctopus/confx/pkg/types/kv"
)

func TestInjection(t *testing.T) {
	var op kv.Executor = &confredis.Endpoint{}

	ctx := kv.Carry(op)(context.Background())

	_, ok := kv.From(ctx)
	testx.Expect(t, ok, testx.BeTrue())
	testx.Expect(t, op, testx.Equal[kv.Executor](kv.Must(ctx)))
}
