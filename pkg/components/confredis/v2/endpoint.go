package confredis

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/redis/go-redis/v9"

	"github.com/xoctopus/confx/pkg/types"
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

	if d := e.LivenessCheck(ctx); !d.Reachable {
		return errors.New(d.Message)
	}

	return nil
}

func (e *Endpoint) LivenessCheck(ctx context.Context) (d types.LivenessData) {
	if e.cli == nil {
		d.Message = "lost connection"
		return
	}

	span := types.Cost()
	cmd := e.cli.Ping(ctx)
	cost := span()

	if err := cmd.Err(); err != nil {
		d.Message = err.Error()
		return
	}

	d.TTL = types.Duration(cost)
	d.Reachable = true
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
