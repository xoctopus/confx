package confredis

import (
	"context"
	"fmt"
	"net/url"

	"github.com/redis/go-redis/v9"
	"github.com/xoctopus/genx/testdata/errors"

	"github.com/xoctopus/confx/pkg/types"
	"github.com/xoctopus/confx/pkg/types/liveness"
)

type Endpoint struct {
	types.Endpoint[Option]

	cli redis.UniversalClient
}

func (e *Endpoint) Init(ctx context.Context) error {
	if err := e.Endpoint.Init(); err != nil {
		return err
	}

	if e.cli != nil {
		return nil
	}

	opt := e.Option.ClientOption()
	for _, addr := range append(opt.Addrs, e.Endpoint.Address) {
		u, err := url.Parse(addr)
		if err != nil {
			return fmt.Errorf("invalid address: %s [cause: %w]", addr, err)
		}
		opt.Addrs = append(opt.Addrs, u.Host)
	}
	// opt.Addrs = slicex.Unique(append(opt.Addrs, e.Endpoint.Endpoint()))
	opt.Username = e.Auth.Username
	opt.Password = e.Auth.Password.String()

	if !e.Cert.IsZero() {
		opt.TLSConfig = e.Cert.Config()
	}

	e.cli = redis.NewUniversalClient(opt)

	d := e.LivenessCheck(ctx)
	return d.FailureReason()
}

func (e *Endpoint) LivenessCheck(ctx context.Context) (d liveness.Result) {
	d = liveness.NewLivenessData()

	if e.cli == nil {
		d.End(errors.New("redis: lost connection"))
		return
	}

	d.End(e.cli.Ping(ctx).Err())
	return
}

func (e *Endpoint) Client() redis.UniversalClient {
	return e.cli
}

func (e *Endpoint) Close() error {
	if e.cli != nil {
		return e.cli.Close()
	}
	return nil
}

func (e *Endpoint) Key(k string) string {
	return e.Option.Prefix + ":" + k
}

func (e *Endpoint) Exec(ctx context.Context, cmd string, args ...any) (any, error) {
	c := e.cli.Do(ctx, append([]any{cmd}, args...)...)
	return c.Result()
}
