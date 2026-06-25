package confjwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xoctopus/x/contextx"

	"github.com/xoctopus/confx/pkg/types"
)

var (
	ErrInvalidSignKey = errors.New("invalid jwt sign key, got empty")
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
		return ErrInvalidSignKey
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
		Expired: c.ExpiresAt(),
		Issuer:  c.Issuer,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString([]byte(c.SignKey))
}

func (c *JWT) GenerateNoExpiration(payload any) (string, error) {
	claim := &Claims{
		Payload: payload,
		Expired: nil,
		Issuer:  c.Issuer,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString([]byte(c.SignKey))
}

func (c *JWT) Parse(v string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(
		v, &Claims{}, func(token *jwt.Token) (any, error) {
			return []byte(c.SignKey), nil
		})

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
	Payload any              `json:"v"`
	Expired *jwt.NumericDate `json:"e,omitempty"`
	Issuer  string           `json:"i,omitempty"`
}

func (c *Claims) GetExpirationTime() (*jwt.NumericDate, error) { return c.Expired, nil }
func (c *Claims) GetIssuedAt() (*jwt.NumericDate, error)       { return nil, nil }
func (c *Claims) GetNotBefore() (*jwt.NumericDate, error)      { return nil, nil }
func (c *Claims) GetIssuer() (string, error)                   { return "", nil }
func (c *Claims) GetSubject() (string, error)                  { return "", nil }
func (c *Claims) GetAudience() (jwt.ClaimStrings, error)       { return nil, nil }

type tCtxJWT struct{}

var (
	JWTFrom  = contextx.From[tCtxJWT, *JWT]
	MustJWT  = contextx.Must[tCtxJWT, *JWT]
	WithJWT  = contextx.With[tCtxJWT, *JWT]
	CarryJWT = contextx.Carry[tCtxJWT, *JWT]
)
