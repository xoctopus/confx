package appx_test

import (
	"os"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/appx"
)

func TestGetRuntime(t *testing.T) {
	cases := []*struct {
		runtime appx.Runtime
		result  appx.Runtime
	}{
		{appx.RUNTIME_PROD, appx.RUNTIME_PROD},
		{appx.RUNTIME_STAGING, appx.RUNTIME_STAGING},
		{appx.RUNTIME_DEV, appx.RUNTIME_DEV},
		{appx.Runtime("invalid"), appx.RUNTIME_PROD},
	}

	for _, c := range cases {
		if err := os.Setenv(appx.RuntimeKey, c.runtime.String()); err != nil {
			t.Fatal(err)
		}

		Expect(t, appx.GetRuntime(), Equal(c.result))

		if err := os.Unsetenv(appx.RuntimeKey); err != nil {
			t.Fatal(err)
		}
	}
}
