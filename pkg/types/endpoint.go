package types

import (
	"context"
	"maps"
	"net"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/xoctopus/x/textx"

	"github.com/xoctopus/confx/pkg/conftls"
	"github.com/xoctopus/confx/pkg/types/liveness"
)

// Endpoint a connectable endpoint
// Note options in url Param can override option
type Endpoint[Option any] struct {
	// Address component connection endpoint address
	Address string
	// ExtraAddress for cluster multi-endpoint
	ExtraAddress []string
	// Auth support Endpoint auth info with username and password
	Auth Userinfo
	// Option component Option. if no option use EndpointNoOption
	Option Option
	// Cert X509KeyPair to support certification info
	Cert conftls.X509KeyPair

	addr *url.URL
}

type EndpointNoOption = Endpoint[struct{}]

func (e *Endpoint[Option]) IsZero() bool {
	return e.Address == ""
}

func (e *Endpoint[Option]) SetDefault() {
	if x, ok := any(&e.Option).(Defaulter); ok {
		x.SetDefault()
	}
	e.Auth.SetDefault()
}

func (e *Endpoint[Option]) Init() (err error) {
	e.addr, err = url.Parse(e.Address)
	if err != nil {
		return
	}

	main := key(e.addr)
	followers := map[string]string{}
	for i := range e.ExtraAddress {
		s := e.ExtraAddress[i]
		u := (*url.URL)(nil)
		if u, err = url.Parse(s); err != nil {
			return
		}

		u = &url.URL{
			Scheme: e.addr.Scheme,
			Host:   u.Host,
			Path:   u.Path,
		}
		ep := key(u)
		if _, ok := followers[ep]; ok || ep == main {
			continue
		}
		followers[ep] = u.String()
	}
	e.ExtraAddress = slices.Collect(maps.Values(followers))

	if e.Auth.IsZero() {
		password, _ := e.addr.User.Password()
		e.Auth.Password = Password(password)
		e.Auth.Username = e.addr.User.Username()
	}
	if err = e.Auth.Init(); err != nil {
		return err
	}
	if !e.Auth.IsZero() {
		e.addr.User = e.Auth.Userinfo()
	}

	if err = textx.UnmarshalURL(e.addr.Query(), &e.Option); err != nil {
		return err
	}
	param, _ := textx.MarshalURL(e.Option)
	if len(param) > 0 {
		e.addr.RawQuery = param.Encode()
	}

	if !e.Cert.IsZero() {
		if err = e.Cert.Init(); err != nil {
			return err
		}
	}

	return nil
}

func key(u *url.URL) string {
	return (&url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   u.Path,
	}).String()
}

// Key returns Scheme, Host and Path. this method helps to identify a unique
// component
func (e *Endpoint[Option]) Key() string {
	return key(e.addr)
}

// String returns full address info includes options and auth info
func (e *Endpoint[Option]) String() string {
	return e.addr.String()
}

// SecurityString like String but auth info is hidden
func (e *Endpoint[Option]) SecurityString() string {
	u := *e.addr
	u.User = nil
	return u.String()
}

func (e *Endpoint[Option]) Scheme() string {
	return e.addr.Scheme
}

func (e *Endpoint[Option]) LivenessCheck(ctx context.Context) (v liveness.Result) {
	v = liveness.NewLivenessData()

	host := e.addr.Host
	if !strings.Contains(host, ":") {
		host += ":80"
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	v.Start()
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", host)
	if conn != nil {
		_ = conn.Close()
	}
	v.End(err)
	return
}

func (e *Endpoint[Option]) URL() url.URL {
	return *e.addr
}

func (e *Endpoint[Option]) AddOption(k string, vs ...string) {
	q := e.addr.Query()
	for _, v := range vs {
		q.Add(k, v)
	}
	e.addr.RawQuery = q.Encode()
}
