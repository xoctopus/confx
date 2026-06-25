package confjwt_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/xoctopus/httpx/pkg/httpx"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/confjwt"
	"github.com/xoctopus/confx/pkg/types"
)

func TestAuthMiddleware(t *testing.T) {
	auth := &confjwt.Auth{
		AuthInQuery:  "",
		AuthInHeader: "",
	}

	conf := &confjwt.JWT{
		Issuer:  "jwt_unit_test",
		ExpIn:   types.Duration(time.Second),
		SignKey: "any",
	}

	ctx := confjwt.WithJWT(context.Background(), conf)

	t.Run("#AuthMiddleware", func(t *testing.T) {
		payload := any("any")
		t.Run("#Output", func(t *testing.T) {
			tok, _ := conf.GenerateNoExpiration(payload)
			t.Run("#InQuery", func(t *testing.T) {
				auth.AuthInHeader = ""
				auth.AuthInQuery = tok
				pl, err := auth.Output(ctx)
				Expect(t, err, Succeed())
				Expect(t, pl, Equal(payload))
			})
			t.Run("#InHeader", func(t *testing.T) {
				auth.AuthInHeader = tok
				auth.AuthInQuery = ""
				pl, err := auth.Output(ctx)
				Expect(t, err, Succeed())
				Expect(t, pl, Equal(payload))
			})
			t.Run("#HasPermissionValidator", func(t *testing.T) {
				t.Run("#Denied", func(t *testing.T) {
					errVa := errors.New("any")
					va := func(context.Context, any) error { return errVa }
					ctx := confjwt.WithPermissionValidator(ctx, va)
					_, err := auth.Output(ctx)
					Expect(t, err, IsError(errVa))
					var se httpx.StatusError
					Expect(t, errors.As(err, &se), BeTrue())
					Expect(t, se.StatusCode(), Equal(http.StatusForbidden))
				})
				t.Run("#Pass", func(t *testing.T) {
					va := func(context.Context, any) error { return nil }
					ctx := confjwt.WithPermissionValidator(ctx, va)
					out, err := auth.Output(ctx)
					Expect(t, err, Succeed())
					Expect(t, out, Equal(payload))

					ctx = context.WithValue(ctx, auth.ContextKey(), out)
					Expect(t, confjwt.AuthorizationFrom(ctx), Equal(payload.(string)))
				})
			})
			t.Run("#InvalidToken", func(t *testing.T) {
				auth.AuthInQuery = "invalid_token"
				_, err := auth.Output(ctx)
				Expect(t, err, Failed())
				var se httpx.StatusError
				Expect(t, errors.As(err, &se), BeTrue())
				Expect(t, se.StatusCode(), Equal(http.StatusUnauthorized))
			})
		})
	})
}
