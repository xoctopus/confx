package confredis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/xoctopus/x/misc/must"

	"github.com/xoctopus/confx/pkg/types"
)

type Endpoint struct {
	types.Endpoint[RedisOptions]

	prefix string
	index  int
	pool   *redis.Pool
}

func (r *Endpoint) SetDefault() {
	if r.Address == "" {
		r.Address = "redis://localhost:6379/0"
	}
}

func (r *Endpoint) Init() error {
	if err := r.Endpoint.Init(); err != nil {
		return err
	}

	r.prefix = r.Option.Prefix
	if r.prefix == "" {
		return errors.New("invalid redis prefix. it must be set")
	}

	index, err := strconv.Atoi(strings.TrimPrefix(r.URL().Path, "/"))
	if err != nil {
		return fmt.Errorf("invalid redis select index, expect an integer, but got %s", r.URL().Path)
	}
	r.index = index

	if r.pool == nil {
		dialer := func() (c redis.Conn, err error) {
			return redis.Dial(
				"tcp", r.URL().Host,
				redis.DialDatabase(r.index),
				redis.DialReadTimeout(time.Duration(r.Option.ReadTimeout)),
				redis.DialWriteTimeout(time.Duration(r.Option.WriteTimeout)),
				redis.DialConnectTimeout(time.Duration(r.Option.ConnTimeout)),
				redis.DialKeepAlive(time.Duration(r.Option.KeepAlive)),
				redis.DialPassword(r.Auth.Password.String()),
				redis.DialTLSSkipVerify(true),
				redis.DialUseTLS(!r.Cert.IsZero()),
			)
		}

		r.pool = &redis.Pool{
			Dial:            dialer,
			MaxConnLifetime: time.Duration(r.Option.KeepAlive),
			MaxIdle:         r.Option.MaxIdle,
			MaxActive:       r.Option.MaxActive,
			IdleTimeout:     time.Duration(r.Option.KeepAlive),
			Wait:            r.Option.Wait,
		}
	}

	if d := r.LivenessCheck(context.Background()); !d.Reachable {
		return errors.New(d.Message)
	}

	return nil
}

func (r *Endpoint) LivenessCheck(_ context.Context) (d types.LivenessData) {
	if conn := r.Get(); conn != nil {
		defer func() { _ = conn.Close() }()

		span := types.Cost()
		_, err := conn.Do("PING")
		if err != nil {
			d.Message = err.Error()
			return
		}
		d.Reachable = true
		d.TTL = types.Duration(span())
		return
	}

	d.Message = "lost connection"

	return
}

func (r *Endpoint) Exec(cmd string, args ...any) (any, error) {
	c := r.Get()
	if c == nil {
		return nil, errors.New("redis: lost connection")
	}
	defer func() { _ = c.Close() }()

	return c.Do(cmd, args...)
}

func (r *Endpoint) Close() error {
	if r.pool != nil {
		return r.pool.Close()
	}
	return nil
}

func (r *Endpoint) Get() redis.Conn {
	if r.pool != nil {
		return r.pool.Get()
	}
	return nil
}

func (r *Endpoint) MustGet() redis.Conn {
	c := r.Get()
	must.BeTrueF(c != nil && c.Err() == nil, "redis: lost connection")
	return c
}

func (r *Endpoint) Key(key string) string {
	return fmt.Sprintf("%s:%s", r.prefix, key)
}
