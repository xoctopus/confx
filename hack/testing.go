package hack

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/xoctopus/logx"
	"github.com/xoctopus/sfid/pkg/sfid"
	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/misc/retry"

	"github.com/xoctopus/confx/pkg/types"
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
	logx.SetLogFormat(logx.LogFormatJSON)

	t.Setenv(types.DEPLOY_ENVIRONMENT, "test_hack")
	t.Setenv(types.TARGET_PROJECT, "test_local")

	return contextx.Compose(
		logx.Carry(logx.NewStd()),
		sfid.Carry(sfid.NewDefaultIDGen(100)),
	)(context.Background())
}
