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

	var payload any
	claims, err := c.Parse(tok)
	if err == nil && claims != nil {
		payload = claims.Payload

		if va, ok := PermissionValidatorFrom(ctx); ok && va != nil {
			err = va(ctx, payload)
		}

		if err != nil {
			return nil, httpx.STATUS__FORBIDDEN.Wrap(err)
		}

		return payload, nil
	}

	return nil, httpx.STATUS__UNAUTHORIZED.Wrap(err)
}

func AuthorizationFrom(ctx context.Context) string {
	return ctx.Value(tCtxAuthorization{}).(string)
}

type (
	PermissionValidator func(context.Context, any) error

	tCtxPermissionValidator struct{}
)

var (
	PermissionValidatorFrom  = contextx.From[tCtxPermissionValidator, PermissionValidator]
	MustPermissionValidator  = contextx.Must[tCtxPermissionValidator, PermissionValidator]
	WithPermissionValidator  = contextx.With[tCtxPermissionValidator, PermissionValidator]
	CarryPermissionValidator = contextx.Carry[tCtxPermissionValidator, PermissionValidator]
)
