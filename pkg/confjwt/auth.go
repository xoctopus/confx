package confjwt

import (
	"context"
	"strings"

	"github.com/xoctopus/httpx/pkg/httpx"
	"github.com/xoctopus/x/contextx"
)

type tCtxAuthorization struct{}

type Auth struct {
	AuthInQuery  string `name:"authorization,omitempty" in:"query"  validate:"@string[1,]"`
	AuthInHeader string `name:"Authorization,omitempty" in:"header" validate:"@string[1,]"`
}

func (r Auth) ContextKey() any {
	return tCtxAuthorization{}
}

func (r Auth) Output(ctx context.Context) (any, error) {
	c := MustJWT(ctx)

	a := r.AuthInQuery
	if a == "" {
		a = r.AuthInHeader
	}

	tok := strings.TrimSpace(strings.Replace(a, "Bearer", " ", 1))

	var (
		payload any
		err     error
		valid   bool
		claims  *Claims
	)

	if parser, ok := BuiltinTokenParserFrom(ctx); ok {
		payload, err, valid = parser(ctx, tok)
	}

	if !valid {
		claims, err = c.Parse(tok)
		if err == nil && claims != nil {
			payload = claims.Payload
		}
	}

	if err != nil {
		return nil, httpx.STATUS__UNAUTHORIZED.Wrap(err)
	}

	if va, ok := PermissionValidatorFrom(ctx); ok && va != nil {
		if err = va(ctx, payload); err != nil {
			return nil, httpx.STATUS__FORBIDDEN.Wrap(err)
		}
	}

	return payload, nil
}

func AuthorizationFrom(ctx context.Context) string {
	return ctx.Value(tCtxAuthorization{}).(string)
}

type (
	BuiltinTokenParser  func(ctx context.Context, tok string) (payload any, err error, valid bool)
	PermissionValidator func(context.Context, any) error

	tCtxPermissionValidator struct{}
	tCtxBuiltinTokenParser  struct{}
)

var (
	PermissionValidatorFrom  = contextx.From[tCtxPermissionValidator, PermissionValidator]
	MustPermissionValidator  = contextx.Must[tCtxPermissionValidator, PermissionValidator]
	WithPermissionValidator  = contextx.With[tCtxPermissionValidator, PermissionValidator]
	CarryPermissionValidator = contextx.Carry[tCtxPermissionValidator, PermissionValidator]

	BuiltinTokenParserFrom  = contextx.From[tCtxBuiltinTokenParser, BuiltinTokenParser]
	MustBuiltinTokenParser  = contextx.Must[tCtxBuiltinTokenParser, BuiltinTokenParser]
	WithBuiltinTokenParser  = contextx.With[tCtxBuiltinTokenParser, BuiltinTokenParser]
	CarryBuiltinTokenParser = contextx.Carry[tCtxBuiltinTokenParser, BuiltinTokenParser]
)
