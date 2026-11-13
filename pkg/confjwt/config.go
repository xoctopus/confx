package confjwt

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xoctopus/x/codex"

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
	return c.generate(&Claims{
		Payload: payload,
		Expired: c.ExpiresAt(),
		Issuer:  c.Issuer,
	})
}

func (c *JWT) GenerateNoExpiration(payload any) (string, error) {
	return c.generate(&Claims{
		Payload: payload,
		Expired: nil,
		Issuer:  c.Issuer,
	})
}

func (c *JWT) generate(claims *Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(c.SignKey))
	if err != nil {
		return "", codex.Wrap(ERROR__TOKEN_GENERATING, err)
	}
	return t, nil
}

func (c *JWT) Parse(v string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(
		v, &Claims{}, func(token *jwt.Token) (any, error) {
			return []byte(c.SignKey), nil
		})

	if err != nil {
		return nil, codex.Wrapf(ERROR__TOKEN_VALIDATION, err, "failed parsing")
	}
	if t == nil {
		return nil, codex.Errorf(ERROR__TOKEN_VALIDATION, "got empty token")
	}
	claim, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, codex.Errorf(ERROR__TOKEN_VALIDATION, "invalid claim")
	}
	return claim, nil
}

func (c *JWT) WithContext(ctx context.Context) context.Context {
	return WithJWT(ctx, c)
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
