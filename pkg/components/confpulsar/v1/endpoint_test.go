package confpulsar_test

import (
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/hack"
	"github.com/xoctopus/confx/pkg/components/confmq"
	"github.com/xoctopus/confx/pkg/components/confpulsar/v1"
)

func TestEndpoint_Hack(t *testing.T) {
	t.Run("SetDefault", func(t *testing.T) {
		ep := &confpulsar.Endpoint{}

		ep.SetDefault()
		Expect(t, ep.Endpoint.String(), Equal("pulsar://localhost:6650"))
	})

	t.Run("Init", func(t *testing.T) {
		t.Run("Lost", func(t *testing.T) {
			ctx := hack.WithPulsarLost(hack.Context(t), t, "pulsar://localhost:6650")
			ep := confmq.Must(ctx)
			Expect(t, ep, NotBeNil[confmq.PubSub]())

			_, err := ep.Subscribe(ctx, "topic")
			Expect(t, err, Failed())

			err = ep.Publish(ctx, confmq.NewMessage(ctx, "topic", "any"))
			Expect(t, err, Failed())
		})

		t.Run("Established", func(t *testing.T) {
			ctx := hack.WithPulsar(hack.Context(t), t, "pulsar://localhost:16650")
			ep := confmq.Must(ctx)
			Expect(t, ep, NotBeNil[confmq.PubSub]())
		})
	})
}
