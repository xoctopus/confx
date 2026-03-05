package confredis_test

import (
	"testing"

	. "cgtech.gitlab.com/saitox/x/testx"
	"github.com/redis/go-redis/v9"

	"cgtech.gitlab.com/saitox/confx/hack"
	"cgtech.gitlab.com/saitox/confx/pkg/confredis"
	"cgtech.gitlab.com/saitox/confx/pkg/types/kv"
)

func TestEndpoint(t *testing.T) {
	t.Run("Init", func(t *testing.T) {
		ctx := hack.WithRedis(
			hack.Context(t), t,
			"redis://:123456@localhost:16379?prefix=hack_test",
		)
		cli := confredis.Must(ctx)
		Expect(t, cli, NotBeNil[confredis.EndpointClient]())
		Expect(t, cli.Client(), NotBeNil[redis.UniversalClient]())

		ep := kv.Must(ctx)
		key := ep.Key("some")
		Expect(t, key, HavePrefix("hack_test:"))

		_, err := ep.Exec(ctx, "set", "k", "v")
		Expect(t, err, Succeed())
		_ = err

		r, err := ep.Exec(ctx, "get", "k")
		Expect(t, err, Succeed())
		Expect(t, r.(string), Equal("v"))

		_, err = ep.Exec(ctx, "del", "k")
		Expect(t, err, Succeed())
	})
}
