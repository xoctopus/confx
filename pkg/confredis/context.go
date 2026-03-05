package confredis

import (
	"context"

	"cgtech.gitlab.com/saitox/x/contextx"
	"cgtech.gitlab.com/saitox/x/misc/must"
	"github.com/redis/go-redis/v9"
)

type EndpointClient interface {
	Client() redis.UniversalClient
}

type tEndpointClient struct{}

func With(ctx context.Context, c EndpointClient) context.Context {
	return context.WithValue(ctx, tEndpointClient{}, c)
}

func From(ctx context.Context) (EndpointClient, bool) {
	c, ok := ctx.Value(tEndpointClient{}).(EndpointClient)
	return c, ok
}

func Must(ctx context.Context) EndpointClient {
	c, ok := From(ctx)
	must.BeTrueF(ok, "missing redis.EndpointClient")
	return c
}

func Carry(c EndpointClient) contextx.Carrier {
	return func(ctx context.Context) context.Context {
		return With(ctx, c)
	}
}
