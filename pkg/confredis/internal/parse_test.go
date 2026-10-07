package internal_test

import (
	"net/url"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/confredis"
	"github.com/xoctopus/confx/pkg/confredis/internal"
)

func TestParseURL(t *testing.T) {
	redisURL, confxQ, err := internal.ParseURL("redis://127.0.0.1:6379/2?prefix=p&max_retries=3&db=1")
	Expect(t, err, Succeed())
	Expect(t, confxQ.Get("prefix"), Equal("p"))
	Expect(t, confxQ.Get("db"), Equal("1"))

	u, err := url.Parse(redisURL)
	Expect(t, err, Succeed())
	Expect(t, u.Query().Get("max_retries"), Equal("3"))
	Expect(t, u.Query().Get("db"), Equal("1"))
	Expect(t, u.Query().Has("prefix"), BeFalse())

	_, _, err = internal.ParseURL("redis://127.0.0.1:6379?ignored=1&max_retries=1")
	Expect(t, err, Succeed())
	opt := confredis.Option{}.ClientOption("redis://127.0.0.1:6379?ignored=1&max_retries=1")
	Expect(t, opt.MaxRetries, Equal(1))
}
