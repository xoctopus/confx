package confotel_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/confotel"
	"github.com/xoctopus/confx/pkg/types"
)

func Setup(t testing.TB, c any) context.Context {
	t.Helper()

	ctx := context.Background()
	err := types.InitByContext(ctx, c)

	testx.Expect(t, err, testx.Be[error](nil))

	t.Cleanup(func() {
		_ = types.CloseByContext(ctx, c)
	})

	if x, ok := c.(types.Injectable); ok {
		return x.WithContext(ctx)
	}
	return ctx
}

func TestLog(t *testing.T) {
	ctx := Setup(t, &confotel.Otel{
		LogLevel:  logx.LogLevelDebug,
		LogFormat: logx.LogFormatTEXT,
	})

	doLog(ctx)
}

func doLog(ctx context.Context) {
	ctx, log := logx.Start(ctx, "op")
	defer log.End()

	otherActions(ctx)
	someActionWithSpan(ctx)
}

func someActionWithSpan(ctx context.Context) {
	_, log := logx.Start(ctx, "SomeActionWithSpan")
	defer log.End()

	log.Info("info msg")
	log.Debug("debug msg")
	log.Warn(errors.New("warn msg"))
}

func otherActions(ctx context.Context) {
	log := logx.From(ctx)

	time.Sleep(200 * time.Millisecond)

	log.With("test2", 2).Info("test")
	log.Error(errors.New("other action failed."))
}
