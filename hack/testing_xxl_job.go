package hack

import (
	"context"
	"net/url"
	"testing"

	"cgtech.gitlab.com/saitox/x/contextx"
	. "cgtech.gitlab.com/saitox/x/testx"

	"cgtech.gitlab.com/saitox/confx/pkg/confxxl"
)

func WithXXLRegistry(ctx context.Context, t testing.TB, dsn string, executors ...string) context.Context {
	_, err := url.Parse(dsn)
	Expect(t, err, Succeed())

	ep := &confxxl.Endpoint{}
	ep.Address = dsn
	ep.Option.Executor = executors
	ep.Option.ClientID = "confx-hack"
	ep.Option.Listener = "http://host.docker.internal:9999"
	ep.Option.AccessToken = "l6MOJuZ12RKzfaM1"
	ep.SetDefault()

	if err = ep.Init(ctx); err != nil {
		return ctx
	}

	t.Cleanup(func() { _ = ep.Close() })
	return contextx.Compose(confxxl.Carry(ep))(ctx)
}
