package confjwt_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/confjwt"
	"github.com/xoctopus/confx/pkg/types"
)

func TestJwtConfig(t *testing.T) {
	conf := &confjwt.JWT{
		Issuer:  "jwt_unit_test",
		ExpIn:   types.Duration(time.Second),
		SignKey: "any",
	}

	t.Run("#JwtConf", func(t *testing.T) {
		t.Run("#Init", func(t *testing.T) {
			c := *conf
			c.SignKey = ""
			Expect(t, c.Init(), IsError(confjwt.ErrInvalidSignKey))

			c.ExpIn = 0
			c.SignKey = "any"
			c.SetDefault()
			Expect(t, c.Init(), Succeed())
			Expect(t, c.ExpIn, Equal(types.Duration(time.Hour)))
		})

		t.Run("#ExpiresAt", func(t *testing.T) {
			c := *conf
			Expect(t, c.ExpiresAt(), NotBeNil[*jwt.NumericDate]())

			c.ExpIn = 0
			Expect(t, c.ExpiresAt(), BeNil[*jwt.NumericDate]())
		})

		t.Run("#GeneratingAndParsing", func(t *testing.T) {
			conf.ExpIn = types.Duration(time.Second * 2)
			payload := any("any")
			tok, err := conf.Generate(payload)
			Expect(t, err, Succeed())
			Expect(t, len(tok) > 0, BeTrue())
			t.Log(tok)

			t.Run("#Success", func(t *testing.T) {
				claim, err := conf.Parse(tok)
				Expect(t, err, Succeed())
				Expect(t, claim.Payload, Equal(payload))
			})

			t.Run("#Failed", func(t *testing.T) {
				t.Run("#TokenExpired", func(t *testing.T) {
					time.Sleep(2 * time.Second)
					_, err = conf.Parse(tok)
					Expect(t, err, Failed())
				})

				t.Run("#ParseWithClaimFailed", func(t *testing.T) {
					_, err = conf.Parse("not equal token gen before")
					Expect(t, err, Failed())
				})
			})
		})
	})
}
