package confredis_test

import (
	"net/url"
	"testing"

	"github.com/xoctopus/x/misc/must"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/hack"
	"github.com/xoctopus/confx/pkg/comp/confredis"
	redisv1 "github.com/xoctopus/confx/pkg/comp/confredis/v1"
	"github.com/xoctopus/confx/pkg/comp/runtime"
	"github.com/xoctopus/confx/pkg/types"
)

func TestEndpoint(t *testing.T) {
	ep := &redisv1.Endpoint{}

	t.Run("SetDefault", func(t *testing.T) {
		ep.SetDefault()
		Expect(t, ep.String(), Equal("redis://127.0.0.1:6379/0"))
		Expect(t, ep.DB(), Equal(0))

	})
	t.Run("Init", func(t *testing.T) {
		t.Run("PrefixFromEnv", func(t *testing.T) {
			t.Setenv(runtime.TARGET_PROJECT, "test")
			t.Setenv(runtime.DEPLOY_ENVIRONMENT, "local")

			Expect(t, ep.Init(), Failed()) // not hacking
			Expect(t, ep.String(), Equal(
				"redis://127.0.0.1:6379/0"+
					"?connTimeout=10"+
					"&keepAlive=3600"+
					"&maxActive=100"+
					"&maxIdle=100"+
					"&readTimeout=10"+
					"&skipTlsVerify=true"+
					"&wait=true"+
					"&writeTimeout=10"))
		})
		t.Run("InvalidParam", func(t *testing.T) {
			ep2 := &redisv1.Endpoint{Endpoint: *must.NoErrorV(types.ParseEndpoint("tcp://localhost:100/1"))}
			ep2.Param = url.Values{"connTimeout": []string{"abc"}}
			Expect(t, ep2.Init(), Failed())
		})
	})
}

func TestEndpoint_Hack(t *testing.T) {
	ctx1 := hack.WithRedis(hack.Context(t), t, "redis://:123456@127.0.0.1:16379/0")
	ctx2 := hack.WithRedisLost(hack.Context(t), t, "redis://:123456@127.0.0.1:16380/0")
	t.Run("Hack", func(t *testing.T) {
		op := confredis.MustFrom(ctx1)
		_, err := op.Exec("set", op.Key("abc"), 1)
		Expect(t, err, Succeed())

		conn := op.(*redisv1.Endpoint).MustGet()
		Expect(t, conn != nil, BeTrue())
		Expect(t, conn.Err() == nil, BeTrue())

		v, _ := op.Exec("get", op.Key("abc"))
		Expect(t, v, Equal[any]([]byte("1")))

		// v, _ = op.ExecCmd(
		// 	Command("get", op.Key("abc")),
		// 	Command("get", op.Key("def")),
		// )
		// Expect(t, v, Equal[any]([]any{[]byte("1"), nil}))

		m := op.(types.LivenessChecker).LivenessCheck()
		Expect(t, m["redis://127.0.0.1:16379/0"], Equal("true"))
	})
	t.Run("Lost", func(t *testing.T) {
		op := confredis.MustFrom(ctx2)
		_, err := op.Exec("set", op.Key("abc"), 1)
		Expect(t, err, Failed())

		m := op.(types.LivenessChecker).LivenessCheck()
		Expect(t, m["redis://127.0.0.1:16380/0"], Equal("false"))
	})
}
