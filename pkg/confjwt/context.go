package confjwt

import "github.com/xoctopus/x/contextx"

type tCtxJWT struct{}

var (
	JWTFrom  = contextx.From[tCtxJWT, *JWT]
	MustJWT  = contextx.Must[tCtxJWT, *JWT]
	WithJWT  = contextx.With[tCtxJWT, *JWT]
	CarryJWT = contextx.Carry[tCtxJWT, *JWT]
)
