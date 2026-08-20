package confredis_test

import (
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/hack"
	"github.com/xoctopus/confx/pkg/confredis"
)

func TestEndpoint(t *testing.T) {
	t.Run("Init", func(t *testing.T) {
		ctx := hack.WithRedis(
			hack.Context(t), t,
			"redis://:123456@localhost:16379",
			func(ep *confredis.Endpoint) { ep.Option.Prefix = "hack_test" },
		)
		cli := confredis.MustClient(ctx)
		Expect(t, cli, NotBeNil[confredis.Client]())

		key := cli.Key("some")
		Expect(t, key, HavePrefix("hack_test:"))

		_, err := cli.Exec(ctx, "set", "k", "v")
		Expect(t, err, Succeed())
		_ = err

		r, err := cli.Exec(ctx, "get", "k")
		Expect(t, err, Succeed())
		Expect(t, r.(string), Equal("v"))

		_, err = cli.Exec(ctx, "del", "k")
		Expect(t, err, Succeed())
	})

	t.Run("InitTLS", func(t *testing.T) {
		ctx := hack.WithRedis(
			hack.Context(t), t,
			"rediss://:123456@localhost:16380",
			func(ep *confredis.Endpoint) { ep.Option.Prefix = "hack_test_tls" },
		)
		cli := confredis.MustClient(ctx)
		Expect(t, cli, NotBeNil[confredis.Client]())

		_, err := cli.Exec(ctx, "ping")
		Expect(t, err, Succeed())
	})
}
