package confredis_test

import (
	"crypto/tls"
	"slices"
	"testing"
	"time"

	. "github.com/xoctopus/x/testx"

	. "github.com/xoctopus/confx/pkg/confredis"
	"github.com/xoctopus/confx/pkg/types"
)

func TestOption_SetDefault(t *testing.T) {
	opt := Option{}
	opt.SetDefault()
	Expect(t, opt, Equal(Option{
		ConnectionTimeout: types.Duration(100 * time.Millisecond),
		OperationTimeout:  types.Duration(100 * time.Millisecond),
		BufferSizeKB:      128,
		PoolSize:          20,
		MaxIdleConnection: 10,
		MaxIdleTime:       types.Duration(time.Hour),
	}))
}

func TestClientOption(t *testing.T) {
	t.Run("standalone redis url", func(t *testing.T) {
		opt := Option{}.ClientOption("redis://user:secret@127.0.0.1:6379/2")
		Expect(t, opt.Addrs, Equal([]string{"127.0.0.1:6379"}))
		Expect(t, opt.Username, Equal("user"))
		Expect(t, opt.Password, Equal("secret"))
		Expect(t, opt.DB, Equal(2))
		Expect(t, opt.IsClusterMode, BeFalse())
		Expect(t, opt.TLSConfig, BeNil[*tls.Config]())
	})

	t.Run("rediss does not set tls in ClientOption", func(t *testing.T) {
		opt := Option{}.ClientOption("rediss://127.0.0.1:6379")
		Expect(t, opt.TLSConfig, BeNil[*tls.Config]())
		Expect(t, opt.Addrs, Equal([]string{"127.0.0.1:6379"}))
	})

	t.Run("cluster mode forces db 0", func(t *testing.T) {
		opt := Option{DB: 1, ClusterMode: true}.
			ClientOption("redis://clustercfg.example.com:6379/1")
		Expect(t, opt.DB, Equal(0))
		Expect(t, opt.IsClusterMode, BeTrue())
		Expect(t, opt.Addrs, Equal([]string{"clustercfg.example.com:6379"}))
	})

	t.Run("others are host ports", func(t *testing.T) {
		opt := Option{}.ClientOption(
			"redis://127.0.0.1:6379",
			"redis://10.0.0.2:6379",
			"10.0.0.4:6379",
		)
		got := slices.Clone(opt.Addrs)
		slices.Sort(got)
		Expect(t, got, Equal([]string{"10.0.0.2:6379", "10.0.0.4:6379", "127.0.0.1:6379"}))
	})

	t.Run("option db wins over url path when set", func(t *testing.T) {
		opt := Option{DB: 3}.ClientOption("redis://127.0.0.1:6379/2")
		Expect(t, opt.DB, Equal(3))
	})

	t.Run("redis advanced query kept option overrides curated", func(t *testing.T) {
		opt := Option{PoolSize: 20, ConnectionTimeout: types.Duration(200 * time.Millisecond)}.
			ClientOption("redis://127.0.0.1:6379/1?max_retries=7&dial_timeout=5s")
		Expect(t, opt.DB, Equal(1))
		Expect(t, opt.MaxRetries, Equal(7))
		Expect(t, opt.PoolSize, Equal(20))
		Expect(t, opt.DialTimeout, Equal(200*time.Millisecond))
	})

	t.Run("unknown redis query panics", func(t *testing.T) {
		defer func() {
			Expect(t, recover(), NotBeNil[any]())
		}()
		_ = Option{}.ClientOption("redis://127.0.0.1:6379?prefix=hack_test")
	})
}

func TestHostPort(t *testing.T) {
	hp, err := HostPort("rediss://u:p@example.com:6380/1")
	Expect(t, err, Succeed())
	Expect(t, hp, Equal("example.com:6380"))

	hp, err = HostPort("127.0.0.1:6379")
	Expect(t, err, Succeed())
	Expect(t, hp, Equal("127.0.0.1:6379"))

	_, err = HostPort("")
	Expect(t, err, Failed())
	_, err = HostPort("not-a-port")
	Expect(t, err, Failed())
}
