package confpulsar_test

import (
	"context"
	"testing"
	"time"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/hack"
	"github.com/xoctopus/confx/pkg/components/confmq"
	"github.com/xoctopus/confx/pkg/components/confpulsar/v1"
	"github.com/xoctopus/confx/pkg/types"
)

func TestEndpoint_Hack(t *testing.T) {
	t.Run("SetDefault", func(t *testing.T) {
		ep := &confpulsar.Endpoint{}

		ep.SetDefault()
		Expect(t, ep.Endpoint.String(), Equal("pulsar://localhost:6650"))
	})

	t.Run("Init", func(t *testing.T) {
		t.Run("Lost", func(t *testing.T) {
			dsn := "pulsar://localhost:6650"

			ctx := hack.WithPulsarLost(hack.Context(t), t, dsn)
			ep := confmq.Must(ctx)
			Expect(t, ep, NotBeNil[confmq.PubSub]())

			_, err := ep.Subscribe(ctx, "topic")
			Expect(t, err, Failed())

			err = ep.Publish(ctx, confmq.NewMessage(ctx, "topic", "any"))
			Expect(t, err, Failed())
		})

		t.Run("Established", func(t *testing.T) {
			dsn := "pulsar://localhost:16650"

			ctx := hack.WithPulsar(hack.Context(t), t, dsn)
			ep := confmq.Must(ctx)
			Expect(t, ep, NotBeNil[confmq.PubSub]())

			msg := confmq.NewMessage(ctx, "liveness", "any")
			ret := make(<-chan error)

			sub, err := ep.Subscribe(ctx, "liveness")
			Expect(t, err, Succeed())

			go func() {
				ret = sub.Run(ctx, func(ctx context.Context, rec confmq.Message) {
					Expect(t, rec.Topic(), Equal(msg.Topic()))
					Expect(t, rec.ID(), Equal(msg.ID()))
					_ = sub.Close()
				})
			}()

			time.Sleep(time.Second)
			err = ep.Publish(ctx, msg, confpulsar.WithPublishSync())
			Expect(t, err, Succeed())
			t.Log(<-ret)

			t.Run("LivenessCheck", func(t *testing.T) {
				c, ok := ep.(types.ComponentLivenessChecker)
				Expect(t, ok, BeTrue())
				res := c.LivenessCheck(ctx)
				t.Log(res[ep.(types.Component)])
			})
		})
	})
}
