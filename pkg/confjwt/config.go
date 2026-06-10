package confjwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xoctopus/x/contextx"

	"github.com/xoctopus/confx/pkg/types"
)

type JWT struct {
	Issuer  string         `url:""`
	ExpIn   types.Duration `url:""`
	SignKey string         `url:""`
}

func (c *JWT) SetDefault() {
	if c.ExpIn == 0 {
		c.ExpIn = types.Duration(time.Hour)
	}
}

func (c *JWT) Init() error {
	if c.SignKey == "" {
		return fmt.Errorf("invalid jwt sign key, got empty")
	}
	return nil
}

func (c *JWT) ExpiresAt() *jwt.NumericDate {
	if c.ExpIn == 0 {
		return nil
	}
	return &jwt.NumericDate{Time: time.Now().UTC().Add(time.Duration(c.ExpIn))}
}

func (c *JWT) Generate(payload any) (string, error) {
	claim := &Claims{
		Payload: payload,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: c.ExpiresAt(),
			Issuer:    c.Issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString([]byte(c.SignKey))
}

func (c *JWT) GenerateNoExpiration(payload any) (string, error) {
	claim := &Claims{
		Payload: payload,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: nil,
			Issuer:    c.Issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString([]byte(c.SignKey))
}

func (c *JWT) Parse(v string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(
		v,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(c.SignKey), nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token, failed to parse: %w", err)
	}
	if t == nil {
		return nil, fmt.Errorf("invalid token, empty token")
	}
	claim, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, fmt.Errorf("invalid token, invalid claim")
	}
	return claim, nil
}

type Claims struct {
	Payload interface{}
	jwt.RegisteredClaims
}

type tCtxJWT struct{}

var (
	JWTv4From  = contextx.From[tCtxJWT, *JWT]
	MustJWTv4  = contextx.Must[tCtxJWT, *JWT]
	WithJWTv4  = contextx.With[tCtxJWT, *JWT]
	CarryJWTv4 = contextx.Carry[tCtxJWT, *JWT]
)
