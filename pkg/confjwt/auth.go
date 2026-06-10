package confjwt

import (
	"context"
	"fmt"
	"strings"

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
	c := MustJWTv4(ctx)

	a := r.AuthInQuery
	if a == "" {
		a = r.AuthInHeader
	}

	tok := strings.TrimSpace(strings.Replace(a, "Bearer", " ", 1))

	var payload any
	if claims, err := c.Parse(tok); err == nil && claims != nil {
		payload = claims.Payload

		if va, ok := PermissionValidatorFrom(ctx); ok && va != nil {
			err = va(payload)
		}

		if err != nil {
			return nil, fmt.Errorf("NoPermission: %w", err)
		}

		return payload, nil
	}

	return nil, fmt.Errorf("NoClaims")
}

func AuthorizationFrom[T any](ctx context.Context) T {
	return ctx.Value(tCtxAuthorization{}).(T)
}

type (
	PermissionValidator func(any) error

	tCtxPermissionValidator struct{}
)

var (
	PermissionValidatorFrom  = contextx.From[tCtxPermissionValidator, PermissionValidator]
	MustPermissionValidator  = contextx.Must[tCtxPermissionValidator, PermissionValidator]
	WithPermissionValidator  = contextx.With[tCtxPermissionValidator, PermissionValidator]
	CarryPermissionValidator = contextx.Carry[tCtxPermissionValidator, PermissionValidator]
)
