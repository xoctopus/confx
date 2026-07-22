package confrabbit_test

import (
	"context"
	"testing"

	"github.com/xoctopus/x/testx/bdd"

	"github.com/xoctopus/confx/hack"
	. "github.com/xoctopus/confx/pkg/confrabbit"
	"github.com/xoctopus/confx/pkg/testdata"
	"github.com/xoctopus/confx/pkg/types"
)

func TestEndpoint(t *testing.T) {
	t.Run("SetDefault", func(t *testing.T) {
		bdd.From(t).Given("empty endpoint", func(t bdd.T) {
			ep := &Endpoint{}
			ep.SetDefault()
			t.Then("should equal default",
				bdd.Equal(ep.Address, "amqp://localhost"),
				bdd.Equal(ep.Option.Vhost, "/"),
				bdd.BeTrue(ep.ResourceManager != nil),
			)
		})
	})

	t.Run("Init", func(t *testing.T) {
		bdd.From(t).Given("single address", func(t bdd.T) {
			ctx := hack.WithRabbitMQ(t.Context(), t, "amqp://guest:guest@localhost:5672")
			_, ok := From(ctx)
			t.Then("should be a readiness client", bdd.BeTrue(ok))
		})

		bdd.From(t).Given("multi-address as cluster", func(t bdd.T) {
			ep := &Endpoint{
				Endpoint: types.Endpoint[Option]{
					Address: "amqp://guest:guest@localhost:5672",
					Option: Option{
						Addresses: []string{
							"amqp://guest:guest@localhost:5672",
							"amqp://guest:guest@localhost:25672",
						},
					},
					Cert: testdata.TLSConfigForTest(),
				},
			}
			ep.SetDefault()
			t.Then("initialized", bdd.Succeed(ep.Init(context.Background())))
		})
	})
}
