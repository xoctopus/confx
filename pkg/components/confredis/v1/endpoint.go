package confredis

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/textx"

	"github.com/xoctopus/confx/pkg/components/conftls"
	"github.com/xoctopus/confx/pkg/components/runtime"
	"github.com/xoctopus/confx/pkg/types"
)

type Endpoint struct {
	types.Endpoint

	prefix string
	index  int
	opt    Options

	pool *redis.Pool
}

type Options struct {
	Prefix        string `url:""`
	ConnTimeout   int    `url:",default=10"`
	WriteTimeout  int    `url:",default=10"`
	ReadTimeout   int    `url:",default=10"`
	KeepAlive     int    `url:",default=3600"`
	MaxActive     int    `url:",default=100"`
	MaxIdle       int    `url:",default=100"`
	ClientName    string `url:""`
	SkipTLSVerify bool   `url:",default=true"`
	Wait          bool   `url:",default=true"`
	conftls.X509KeyPair
}

func (r *Endpoint) LivenessCheck() map[string]string {
	k := r.Endpoint.Key()
	m := map[string]string{k: "false"}

	if conn := r.Get(); conn != nil {
		defer func() { _ = conn.Close() }()
		_, err := conn.Do("PING")
		m[k] = fmt.Sprint(err == nil)
	}

	return m
}

func (r *Endpoint) SetDefault() {
	if r.Endpoint.IsZero() {
		must.NoError(r.Endpoint.UnmarshalText([]byte("redis://127.0.0.1:6379/0")))
	}
	r.index = must.NoErrorV(strconv.Atoi(r.Endpoint.Base))
}

func (r *Endpoint) DB() int {
	return r.index
}

func (r *Endpoint) Init() error {
	if err := textx.UnmarshalURL(r.Endpoint.Param, &r.opt); err != nil {
		return err
	}
	param := must.NoErrorV(textx.MarshalURL(r.opt))
	r.Endpoint.Param = param

	r.prefix = r.opt.Prefix
	if r.prefix == "" {
		parts := make([]string, 0, 2)
		project := r.Endpoint.Param.Get("project")
		if project == "" {
			project = strings.ToLower(os.Getenv(runtime.TARGET_PROJECT))
		}
		must.BeTrueF(project != "", "redis project must be set")
		parts = append(parts, project)

		deploy := r.Endpoint.Param.Get("deploy")
		if deploy == "" {
			deploy = strings.ToLower(os.Getenv(runtime.DEPLOY_ENVIRONMENT))
		}
		if deploy != "" {
			parts = append(parts, deploy)
		}
		r.prefix = strings.Join(parts, ":") + ":"
	}

	if r.pool == nil {
		dialer := func() (c redis.Conn, err error) {
			return redis.Dial(
				"tcp", r.Endpoint.Hostname(),
				redis.DialDatabase(r.index),
				redis.DialConnectTimeout(time.Duration(r.opt.ConnTimeout)*time.Second),
				redis.DialWriteTimeout(time.Duration(r.opt.WriteTimeout)*time.Second),
				redis.DialReadTimeout(time.Duration(r.opt.ReadTimeout)*time.Second),
				redis.DialKeepAlive(time.Duration(r.opt.KeepAlive)*time.Second),
				redis.DialPassword(r.Endpoint.Password.String()),
				redis.DialTLSSkipVerify(true),
				redis.DialUseTLS(!r.opt.X509KeyPair.IsZero()),
			)
		}

		r.pool = &redis.Pool{
			Dial:            dialer,
			MaxConnLifetime: time.Duration(r.opt.KeepAlive) * time.Second,
			MaxIdle:         r.opt.MaxIdle,
			MaxActive:       r.opt.MaxActive,
			IdleTimeout:     time.Duration(r.opt.KeepAlive) * time.Second,
			Wait:            r.opt.Wait,
		}
	}
	if _, err := r.Exec("PING"); err != nil {
		_ = r.pool.Close()
		r.pool = nil
		return err
	}

	return nil
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
