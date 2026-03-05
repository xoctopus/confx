package confredis_test

import (
	"testing"
	"time"

	"cgtech.gitlab.com/saitox/x/testx"

	"cgtech.gitlab.com/saitox/confx/pkg/confredis"
	"cgtech.gitlab.com/saitox/confx/pkg/types"
)

func TestOption_SetDefault(t *testing.T) {
	opt := confredis.Option{}
	opt.SetDefault()
	testx.Expect(t, opt, testx.Equal(confredis.Option{
		ConnectionTimeout: types.Duration(100 * time.Millisecond),
		OperationTimeout:  types.Duration(100 * time.Millisecond),
		BufferSizeKB:      128,
		PoolSize:          20,
		MaxIdleConnection: 10,
		MaxIdleTime:       types.Duration(time.Hour),
	}))
}
