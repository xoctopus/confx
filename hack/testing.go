package hack

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/xoctopus/confx/pkg/comp/confredis"
	"github.com/xoctopus/confx/pkg/comp/runtime"
	"github.com/xoctopus/logx"
	. "github.com/xoctopus/x/testx"
)

var once sync.Once

func Check(t testing.TB) {
	if os.Getenv("HACK_TEST") != "true" {
		t.Skip("HACK_TEST=false skip hack testing")
	}
	once.Do(func() {
		fmt.Println("HACK_TEST ENABLED")
		time.Sleep(time.Second * 2) // to wait dependencies ready
	})
}

func Context(t testing.TB) context.Context {
	t.Helper()

	t.Setenv(runtime.DEPLOY_ENVIRONMENT, "test_hack")
	t.Setenv(runtime.TARGET_PROJECT, "test_local")

	return logx.With(context.Background(), logx.Std(logx.NewHandler()))
}

func WithRedis(ctx context.Context, t testing.TB, dsn string) context.Context {
	Check(t)

	_, err := url.Parse(dsn)
	Expect(t, err, Succeed())

	ep := &confredis.Endpoint{}
	Expect(t, ep.UnmarshalText([]byte(dsn)), Succeed())
	Expect(t, ep.Init(), Succeed())

	t.Cleanup(func() {
		_ = ep.Close()
	})

	return confredis.With(ctx, ep)
}

func WithRedisLost(ctx context.Context, t testing.TB, dsn string) context.Context {
	_, err := url.Parse(dsn)
	Expect(t, err, Succeed())

	ep := &confredis.Endpoint{}
	Expect(t, ep.UnmarshalText([]byte(dsn)), Succeed())
	Expect(t, ep.Init(), Failed())

	t.Cleanup(func() {
		_ = ep.Close()
	})

	return confredis.With(ctx, ep)
}
