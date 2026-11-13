package confjwt_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/xoctopus/x/codex"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/confjwt"
	"github.com/xoctopus/confx/pkg/types"
)

func TestAuthMiddleware(t *testing.T) {
	auth := &confjwt.Auth{}

	conf := &confjwt.JWT{
		Issuer:  "jwt_unit_test",
		ExpIn:   types.Duration(time.Second),
		SignKey: "any",
	}

	ctx := confjwt.WithJWT(context.Background(), conf)
	payload := any("any")
	tok, err := conf.GenerateNoExpiration(payload)
	Expect(t, err, Succeed())

	t.Run("#InQuery", func(t *testing.T) {
		auth.AuthInHeader = ""
		auth.AuthInQuery = tok
		pl, err := auth.Output(ctx)
		Expect(t, err, Succeed())
		Expect(t, pl, Equal(payload))
	})

	t.Run("#InHeader", func(t *testing.T) {
		auth.AuthInHeader = "Bearer " + tok
		auth.AuthInQuery = ""
		pl, err := auth.Output(ctx)
		Expect(t, err, Succeed())
		Expect(t, pl, Equal(payload))
	})

	t.Run("#AuthorizationFrom", func(t *testing.T) {
		auth.AuthInHeader = tok
		auth.AuthInQuery = ""
		out, err := auth.Output(ctx)
		Expect(t, err, Succeed())

		ctx = context.WithValue(ctx, auth.ContextKey(), out)
		Expect(t, confjwt.AuthorizationFrom(ctx), Equal(payload.(string)))
		v, _ := confjwt.MayAuthorizationFrom(ctx)
		Expect(t, v, Equal(payload.(string)))
	})

	t.Run("#EmptyToken", func(t *testing.T) {
		auth.AuthInQuery = "Bearer "
		auth.AuthInHeader = ""
		_, err := auth.Output(ctx)
		Expect(t, err, IsCodeError(confjwt.ERROR__TOKEN_VALIDATION))
		Expect(t, statusCode(err), Equal(http.StatusUnauthorized))
	})

	t.Run("#InvalidToken", func(t *testing.T) {
		auth.AuthInQuery = "invalid_token"
		auth.AuthInHeader = ""
		_, err := auth.Output(ctx)
		Expect(t, err, IsCodeError(confjwt.ERROR__TOKEN_VALIDATION))
		Expect(t, statusCode(err), Equal(http.StatusUnauthorized))
	})
}

func statusCode(err error) int {
	code, ok := codex.AsCode[confjwt.Error](err)
	if !ok {
		return 0
	}
	return code.StatusCode()
}
