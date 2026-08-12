package confredis

import (
	"github.com/redis/go-redis/v9"
	"github.com/xoctopus/x/contextx"

	"github.com/xoctopus/confx/pkg/types/kv"
)

// Client redis.UniversalClient + kv.Executor
type Client interface {
	redis.UniversalClient
	kv.Executor
}

type tCtxClient struct{}

var (
	ClientFrom  = contextx.From[tCtxClient, Client]
	WithClient  = contextx.With[tCtxClient, Client]
	MustClient  = contextx.Must[tCtxClient, Client]
	CarryClient = contextx.Carry[tCtxClient, Client]
)
