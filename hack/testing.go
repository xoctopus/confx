package hack

import (
	"context"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/xoctopus/logx"
	"github.com/xoctopus/logx/handlers"
	"github.com/xoctopus/sfid/pkg/sfid"
	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/misc/retry"
	. "github.com/xoctopus/x/testx"

	pulsarv1 "github.com/xoctopus/confx/pkg/confpulsar/v1"
	redisv1 "github.com/xoctopus/confx/pkg/confredis/v1"
	"github.com/xoctopus/confx/pkg/confredis/v2"
	"github.com/xoctopus/confx/pkg/types"
	"github.com/xoctopus/confx/pkg/types/kv"
	"github.com/xoctopus/confx/pkg/types/mq"
)

var retrier = &retry.Retry{
	Repeats:  10,
	Interval: 3 * time.Second,
}

func Check(t testing.TB) {
	if os.Getenv("HACK_TEST") != "true" {
		t.Skip("HACK_TEST=false skip hack testing")
	}
}

func Context(t testing.TB) context.Context {
	t.Helper()
	handlers.SetLogFormat(handlers.LogFormatJSON)

	t.Setenv(types.DEPLOY_ENVIRONMENT, "test_hack")
	t.Setenv(types.TARGET_PROJECT, "test_local")

	return contextx.Compose(
		logx.Carry(logx.NewStd()),
		sfid.Carry(sfid.NewDefaultIDGen(100)),
	)(context.Background())
}

func WithRedis(ctx context.Context, t testing.TB, dsn string) context.Context {
	Check(t)

	_, err := url.Parse(dsn)
	Expect(t, err, Succeed())

	ep := &redisv1.Endpoint{}
	ep.Address = dsn

	err = retrier.Do(func() error { return ep.Init() })
	Expect(t, err, Succeed())

	t.Cleanup(func() { _ = ep.Close() })

	return kv.With(ctx, ep)
}

func WithRedisLost(ctx context.Context, t testing.TB, dsn string) context.Context {
	_, err := url.Parse(dsn)
	Expect(t, err, Succeed())

	ep := &redisv1.Endpoint{}
	ep.Address = dsn

	Expect(t, ep.Init(), Failed())

	t.Cleanup(func() { _ = ep.Close() })

	return kv.Carry(ep)(ctx)
}

func WithRedisV2(ctx context.Context, t testing.TB, dsn string) context.Context {
	Check(t)

	_, err := url.Parse(dsn)
	Expect(t, err, Succeed())

	ep := &confredis.Endpoint{}
	ep.Address = dsn

	err = retrier.Do(func() error { return ep.Init(ctx) })
	Expect(t, err, Succeed())

	t.Cleanup(func() { _ = ep.Close() })

	return confredis.Carry(ep)(ctx)
}

func WithPulsar(ctx context.Context, t testing.TB, dsn string) context.Context {
	Check(t)

	_, err := url.Parse(dsn)
	Expect(t, err, Succeed())

	ep := &pulsarv1.Endpoint{}
	ep.Address = dsn
	ep.SetDefault()

	err = retrier.Do(func() error { return ep.Init(ctx) })
	Expect(t, err, Succeed())

	t.Cleanup(func() {
		_ = ep.Close()
	})
	return mq.With(ctx, ep)
}

func TryWithPulsar(ctx context.Context, t testing.TB, dsn string) (context.Context, error) {
	Check(t)

	_, err := url.Parse(dsn)
	Expect(t, err, Succeed())

	ep := &pulsarv1.Endpoint{}
	ep.Address = dsn
	ep.SetDefault()

	err = ep.Init(ctx)
	if err == nil {
		t.Cleanup(func() { _ = ep.Close() })
		return mq.With(ctx, ep), nil
	}
	return ctx, err
}

func WithPulsarLost(ctx context.Context, t testing.TB, dsn string) context.Context {
	_, err := url.Parse(dsn)
	Expect(t, err, Succeed())

	ep := &pulsarv1.Endpoint{}
	ep.SetDefault()
	ep.Address = dsn

	Expect(t, ep.Init(ctx), Failed())

	t.Cleanup(func() { _ = ep.Close() })

	return mq.Carry(ep)(ctx)
}
