package confredis_test

import (
	"context"
	"testing"

	"github.com/gomodule/redigo/redis"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/hack"
	"github.com/xoctopus/confx/pkg/confredis/v1"
	"github.com/xoctopus/confx/pkg/types"
	"github.com/xoctopus/confx/pkg/types/kv"
)

func TestRedisEndpointV1(t *testing.T) {
	t.Run("SetDefault", func(t *testing.T) {
		ep := &confredis.Endpoint{}

		t.Run("BeforeInit", func(t *testing.T) {
			d := ep.LivenessCheck(context.Background())
			Expect(t, d.Message, Equal("lost connection"))
			Expect(t, ep.Get(), BeNil[redis.Conn]())
			_, err := ep.Exec("any")
			Expect(t, err, ErrorContains("lost connection"))
			Expect(t, ep.Close(), Succeed())
		})

		ep.SetDefault()
		_ = ep.Endpoint.Init()
		Expect(t, ep.Endpoint.Endpoint(), Equal("redis://localhost:6379/0"))
	})

	t.Run("Lost", func(t *testing.T) {
		t.Run("InvalidAddress", func(t *testing.T) {
			ep := &confredis.Endpoint{}
			ep.Address = "redis://localhost:6379/%zz"
			Expect(t, ep.Init(), Failed())
		})

		t.Run("InvalidPrefix", func(t *testing.T) {
			ep := &confredis.Endpoint{}
			ep.SetDefault()
			Expect(t, ep.Init(), ErrorContains("invalid redis prefix"))
		})
		t.Run("InvalidPath", func(t *testing.T) {
			ep := &confredis.Endpoint{}
			ep.Address = "redis://localhost:6379/abc?prefix=unittest"
			Expect(t, ep.Init(), ErrorContains("invalid redis select index"))
		})
		ep := &confredis.Endpoint{}
		ep.Address = "redis://localhost:6379/1?prefix=unittest"
		Expect(t, ep.Init(), Failed())

		_, err := ep.Exec("set", ep.Key("abc"), 1)
		Expect(t, err, Failed())

		d := ep.LivenessCheck(context.Background())
		Expect(t, d.Reachable, BeFalse())
	})

	t.Run("Established", func(t *testing.T) {
		ctx1 := hack.WithRedis(hack.Context(t), t, "redis://:123456@127.0.0.1:16379/0?prefix=unittest")

		ep := kv.Must(ctx1)
		_, err := ep.Exec("set", ep.Key("abc"), 1)
		Expect(t, err, Succeed())

		conn := ep.(*confredis.Endpoint).MustGet()
		Expect(t, conn != nil, BeTrue())
		Expect(t, conn.Err() == nil, BeTrue())

		v, _ := ep.Exec("get", ep.Key("abc"))
		Expect(t, v, Equal[any]([]byte("1")))

		d := ep.(types.LivenessChecker).LivenessCheck(ctx1)
		Expect(t, d.Reachable, BeTrue())
	})
}
