package confkv_test

import (
	"context"
	"testing"

	"github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/confkv"
	"github.com/xoctopus/confx/pkg/confredis/v1"
)

func TestInjection(t *testing.T) {
	var op confkv.Executor = &confredis.Endpoint{}

	ctx := confkv.Carry(op)(context.Background())

	_, ok := confkv.From(ctx)
	testx.Expect(t, ok, testx.BeTrue())
	testx.Expect(t, op, testx.Equal[confkv.Executor](confkv.Must(ctx)))
}
