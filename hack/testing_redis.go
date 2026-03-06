package hack

import (
	"context"
	"net/url"
	"testing"

	"github.com/xoctopus/x/contextx"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/confredis"
	"github.com/xoctopus/confx/pkg/types/kv"
)

func WithRedisLost(ctx context.Context, t testing.TB, dsn string) context.Context {
	_, err := url.Parse(dsn)
	Expect(t, err, Succeed())

	ep := &confredis.Endpoint{}
	ep.Address = dsn

	Expect(t, ep.Init(ctx), Failed())

	t.Cleanup(func() { _ = ep.Close() })

	return contextx.Compose(
		confredis.Carry(ep),
		kv.Carry(ep),
	)(ctx)
}

func WithRedis(ctx context.Context, t testing.TB, dsn string) context.Context {
	Check(t)

	_, err := url.Parse(dsn)
	Expect(t, err, Succeed())

	ep := &confredis.Endpoint{}
	ep.Address = dsn

	err = retrier.Do(func() error { return ep.Init(ctx) })
	Expect(t, err, Succeed())

	t.Cleanup(func() { _ = ep.Close() })

	return contextx.Compose(
		confredis.Carry(ep),
		kv.Carry(ep),
	)(ctx)
}
