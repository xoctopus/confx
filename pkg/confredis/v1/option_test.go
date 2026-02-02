package confredis_test

import (
	"testing"
	"time"

	"github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/confredis/v1"
	"github.com/xoctopus/confx/pkg/types"
)

func TestOption_SetDefault(t *testing.T) {
	opt := &confredis.Option{}
	opt.SetDefault()
	testx.Expect(t, *opt, testx.Equal(confredis.Option{
		ConnTimeout:   types.Duration(10 * time.Second),
		WriteTimeout:  types.Duration(10 * time.Second),
		ReadTimeout:   types.Duration(10 * time.Second),
		KeepAlive:     types.Duration(1 * time.Hour),
		MaxActive:     100,
		MaxIdle:       100,
		SkipTLSVerify: true,
		Wait:          true,
	}))
}
