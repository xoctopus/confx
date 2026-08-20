package confws

import (
	"github.com/xoctopus/x/contextx"
)

type tCtxEndpoint struct{}

var (
	EndpointFrom  = contextx.From[tCtxEndpoint, *Endpoint]
	MustEndpoint  = contextx.Must[tCtxEndpoint, *Endpoint]
	WithEndpoint  = contextx.With[tCtxEndpoint, *Endpoint]
	CarryEndpoint = contextx.Carry[tCtxEndpoint, *Endpoint]
)

type tCtxSession struct{}

type Session interface {
	Session(id string) Client
}

var (
	SessionFrom  = contextx.From[tCtxSession, Session]
	WithSession  = contextx.With[tCtxSession, Session]
	MustSession  = contextx.Must[tCtxSession, Session]
	CarrySession = contextx.Carry[tCtxSession, Session]
)
