package confredis_test

import (
	"testing"

	"github.com/xoctopus/confx/pkg/confredis/v2"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/hack"
)

func TestEndpoint(t *testing.T) {
	t.Run("Init", func(t *testing.T) {
		t.Setenv("HACK_TEST", "true")
		ctx := hack.WithRedisV2(
			hack.Context(t), t,
			"redis://:123456@localhost:16379?prefix=hack_test",
		)
		cli := confredis.Must(ctx)
		Expect(t, cli, NotBeNil[confredis.EndpointClient]())
	})
}
