package hack

import (
	"context"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/xoctopus/logx"
	"github.com/xoctopus/sfid/pkg/sfid"
	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/misc/retry"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/components/confmq"
	pulsarv1 "github.com/xoctopus/confx/pkg/components/confpulsar/v1"
	"github.com/xoctopus/confx/pkg/components/confredis"
	redisv1 "github.com/xoctopus/confx/pkg/components/confredis/v1"
	"github.com/xoctopus/confx/pkg/components/runtime"
	"github.com/xoctopus/confx/pkg/types"
)

var once sync.Once

func Check(t testing.TB, deps ...types.LivenessChecker) {
	if os.Getenv("HACK_TEST") != "true" {
		t.Skip("HACK_TEST=false skip hack testing")
	}
}

func Context(t testing.TB) context.Context {
	t.Helper()

	t.Setenv(runtime.DEPLOY_ENVIRONMENT, "test_hack")
	t.Setenv(runtime.TARGET_PROJECT, "test_local")

	return contextx.Compose(
		logx.Carry(logx.DefaultStd()),
		sfid.Carry(sfid.NewDefaultIDGen(100)),
	)(context.Background())
}

func WithRedis(ctx context.Context, t testing.TB, dsn string) context.Context {
	Check(t)

	_, err := url.Parse(dsn)
	Expect(t, err, Succeed())

	ep := &redisv1.Endpoint{}
	Expect(t, ep.UnmarshalText([]byte(dsn)), Succeed())

	err = (&retry.Retry{
		Repeats:  3,
		Interval: time.Second * 3,
	}).Do(func() error {
		return ep.Init()
	})
	Expect(t, err, Succeed())

	t.Cleanup(func() { _ = ep.Close() })

	return confredis.With(ctx, ep)
}

func WithRedisLost(ctx context.Context, t testing.TB, dsn string) context.Context {
	_, err := url.Parse(dsn)
	Expect(t, err, Succeed())

	ep := &redisv1.Endpoint{}
	Expect(t, ep.UnmarshalText([]byte(dsn)), Succeed())
	Expect(t, ep.Init(), Failed())

	t.Cleanup(func() {
		_ = ep.Close()
	})

	return confredis.Carry(ep)(ctx)
}

func WithPulsar(ctx context.Context, t testing.TB, dsn string) context.Context {
	Check(t)

	_, err := url.Parse(dsn)
	Expect(t, err, Succeed())

	ep := &pulsarv1.Endpoint{}
	Expect(t, ep.UnmarshalText([]byte(dsn)), Succeed())
	ep.SetDefault()

	err = (&retry.Retry{
		Repeats:  3,
		Interval: time.Second * 3,
	}).Do(func() error {
		return ep.Init(ctx)
	})
	Expect(t, err, Succeed())

	t.Cleanup(func() { _ = ep.Close() })
	return confmq.With(ctx, ep)
}

func WithPulsarLost(ctx context.Context, t testing.TB, dsn string) context.Context {
	_, err := url.Parse(dsn)
	Expect(t, err, Succeed())

	ep := &pulsarv1.Endpoint{}
	Expect(t, ep.UnmarshalText([]byte(dsn)), Succeed())

	ep.SetDefault()
	Expect(t, ep.Init(ctx), Failed())

	t.Cleanup(func() {
		_ = ep.Close()
	})

	return confmq.Carry(ep)(ctx)
}
