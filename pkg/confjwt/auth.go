package confjwt

import (
	"context"
	"strings"

	"github.com/xoctopus/x/codex"
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
	if len(tok) == 0 {
		return nil, codex.Errorf(ERROR__TOKEN_VALIDATION, "got empty token")
	}

	claims, err := c.Parse(tok)
	if err != nil {
		return nil, err
	}

	return claims.Payload, nil
}

func AuthorizationFrom(ctx context.Context) string {
	return ctx.Value(tCtxAuthorization{}).(string)
}

func MayAuthorizationFrom(ctx context.Context) (string, bool) {
	payload, ok := ctx.Value(tCtxAuthorization{}).(string)
	return payload, ok
}
