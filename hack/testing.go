package hack

import (
	"context"
	"os"
	"testing"
	"time"

	"cgtech.gitlab.com/saitox/logx"
	"cgtech.gitlab.com/saitox/sfid/pkg/sfid"
	"cgtech.gitlab.com/saitox/x/contextx"
	"cgtech.gitlab.com/saitox/x/misc/retry"

	"cgtech.gitlab.com/saitox/confx/pkg/types"
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
