package confapp_test

import (
	"os"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/confapp"
)

func TestGetRuntime(t *testing.T) {
	cases := []*struct {
		runtime confapp.Runtime
		result  confapp.Runtime
	}{
		{confapp.RUNTIME_PROD, confapp.RUNTIME_PROD},
		{confapp.RUNTIME_STAGING, confapp.RUNTIME_STAGING},
		{confapp.RUNTIME_DEV, confapp.RUNTIME_DEV},
		{confapp.Runtime("invalid"), confapp.RUNTIME_PROD},
	}

	for _, c := range cases {
		if err := os.Setenv(confapp.RuntimeKey, c.runtime.String()); err != nil {
			t.Fatal(err)
		}

		Expect(t, confapp.GetRuntime(), Equal(c.result))

		if err := os.Unsetenv(confapp.RuntimeKey); err != nil {
			t.Fatal(err)
		}
	}
}
