package confredis

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/xoctopus/confx/pkg/types"
	"github.com/xoctopus/confx/pkg/types/kv"
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

func (e *Endpoint) Close() error {
	if e.cli != nil {
		return e.cli.Close()
	}
	return nil
}

var (
	_ kv.Executor      = (*Endpoint)(nil)
	_ kv.Store         = (*Endpoint)(nil)
	_ types.Injectable = (*Endpoint)(nil)
)

func (e *Endpoint) Key(k string) string {
	return e.Option.Prefix + ":" + k
}

func (e *Endpoint) Exec(ctx context.Context, cmd string, args ...any) (any, error) {
	c := e.cli.Do(ctx, append([]any{cmd}, args...)...)
	return c.Result()
}

func (e *Endpoint) Get(ctx context.Context, key string) (string, bool, error) {
	val, err := e.cli.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", false, nil
		}
		return "", false, err
	}
	return val, true, nil
}

func (e *Endpoint) Set(ctx context.Context, key, val string, ttl time.Duration) error {
	if ttl < 0 {
		ttl = 0
	}
	return e.cli.Set(ctx, key, val, ttl).Err()
}

func (e *Endpoint) SetNX(ctx context.Context, key, val string, ttl time.Duration) (bool, error) {
	if ttl < 0 {
		ttl = 0
	}
	return e.cli.SetNX(ctx, key, val, ttl).Result()
}

func (e *Endpoint) Del(ctx context.Context, key string) (bool, error) {
	n, err := e.cli.Del(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (e *Endpoint) TTL(ctx context.Context, key string) (time.Duration, bool, error) {
	d, err := e.cli.PTTL(ctx, key).Result()
	if err != nil {
		return 0, false, err
	}
	switch d {
	case -2: // key does not exist
		return 0, false, nil
	case -1: // key exists but has no expiration
		return 0, true, nil
	default:
		return d, true, nil
	}
}

func (e *Endpoint) WithContext(ctx context.Context) context.Context {
	x := struct {
		redis.UniversalClient
		kv.Executor
	}{
		UniversalClient: e.cli,
		Executor:        e,
	}

	return WithClient(ctx, x)
}
