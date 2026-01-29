package types_test

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/components/conftls"
	"github.com/xoctopus/confx/pkg/envx"
	"github.com/xoctopus/confx/pkg/types"
)

type MockOption struct {
	Timeout types.Duration `url:"timeout,default=10s"`
	Name    string         `url:"name,default='default'"`
}

var DefaultMockOption = MockOption{
	Timeout: types.Duration(10 * time.Second),
	Name:    "default",
}

func (o MockOption) IsZero() bool {
	return o == MockOption{} || o == DefaultMockOption
}

func (o *MockOption) SetDefault() {
	o.Timeout = types.Duration(time.Second * 10)
	o.Name = "unknown"
}

func ExampleEndpoint() {
	grp := envx.NewGroup("TEST")
	enc := envx.NewEncoder(grp)
	err := enc.Encode(types.Endpoint[MockOption]{
		Address: "redis://localhost:6379/1",
		Auth:    types.Userinfo{Username: "username", Password: "LelzsnHN2xnJd/MB+JGIXWqd8pJPhPYfuRfDbrCsZE8="},
		Option:  MockOption{},
		Cert: conftls.X509KeyPair{
			Key: "key_path",
			Crt: "crt_path",
			CA:  "ca_path",
		},
	})
	if err != nil {
		return
	}

	// for configuration
	fmt.Println(string(grp.Bytes()))
	fmt.Println(string(grp.MaskBytes()))

	// Output:
	// TEST__Address=redis://localhost:6379/1
	// TEST__Auth_DecryptKeyEnv=
	// TEST__Auth_Password=LelzsnHN2xnJd/MB+JGIXWqd8pJPhPYfuRfDbrCsZE8=
	// TEST__Auth_Username=username
	// TEST__Cert_CA=ca_path
	// TEST__Cert_Crt=crt_path
	// TEST__Cert_Key=key_path
	// TEST__Option_Name=
	// TEST__Option_Timeout=0s
	//
	// TEST__Address=redis://localhost:6379/1
	// TEST__Auth_DecryptKeyEnv=
	// TEST__Auth_Password=--------
	// TEST__Auth_Username=username
	// TEST__Cert_CA=ca_path
	// TEST__Cert_Crt=crt_path
	// TEST__Cert_Key=key_path
	// TEST__Option_Name=
	// TEST__Option_Timeout=0s
}

func TestEndpoint(t *testing.T) {
	type Endpoint = types.Endpoint[MockOption]

	t.Run("IsZero", func(t *testing.T) {
		ep := &types.EndpointNoOption{Address: ""}
		Expect(t, ep.IsZero(), BeTrue())
		ep = &types.EndpointNoOption{Address: "https://abc.def.com"}
		Expect(t, ep.IsZero(), BeFalse())
	})

	t.Run("InvalidAddress", func(t *testing.T) {
		ep := &Endpoint{Address: "https://example.com/%zz"}
		Expect(t, ep.Init(), Failed())
	})
	t.Run("InvalidAuth", func(t *testing.T) {
		t.Setenv("PASSWORD_DEC_KEY", "def")
		ep := &Endpoint{
			Address: "redis://localhost:6379/1",
			Auth: types.Userinfo{
				Username: "username",
				Password: types.Password(base64.StdEncoding.EncodeToString([]byte("abc"))),
			},
		}
		ep.SetDefault()
		Expect(t, ep.Init(), Failed())

		t.Run("UserinfoInURL", func(t *testing.T) {
			username := ep.Auth.Username
			password := ep.Auth.Password
			ep = &Endpoint{
				Address: fmt.Sprintf("redis://%s:%s@localhost:6379/1", username, password),
				Auth:    types.Userinfo{},
			}
			ep.SetDefault()
			Expect(t, ep.Init(), Failed())
		})
	})
	t.Run("FailedUnmarshalOption", func(t *testing.T) {
		ep := &Endpoint{
			Address: "redis://localhost:6379/1?timeout=abc",
		}
		Expect(t, ep.Init(), Failed())
	})
	t.Run("FailedInitCert", func(t *testing.T) {
		ep := &Endpoint{
			Address: "redis://localhost:6379/1?timeout=3s&name=abc",
			Cert: conftls.X509KeyPair{
				Key: "key_path",
				Crt: "crt_path",
				CA:  "ca_path",
			},
		}
		Expect(t, ep.Init(), Failed())
		Expect(t, ep.Option.Name, Equal("abc"))
		Expect(t, ep.Option.Timeout, Equal(types.Duration(3*time.Second)))
	})

	t.Run("Success", func(t *testing.T) {
		t.Run("URLQuery", func(t *testing.T) {
			ep := &Endpoint{
				Address: "redis://username:password@localhost:6379/1?timeout=3s&name=abc",
			}
			Expect(t, ep.Init(), Succeed())
			Expect(t, ep.Endpoint(), Equal("redis://localhost:6379/1"))
			Expect(t, ep.String(), Equal("redis://username:password@localhost:6379/1?name=abc&timeout=3s"))
			Expect(t, ep.SecurityString(), Equal("redis://localhost:6379/1?name=abc&timeout=3s"))
		})
		t.Run("URLQueryOverrideOption", func(t *testing.T) {
			ep := &Endpoint{
				Address: "redis://username:password@localhost:6379/1?timeout=3s&name=abc",
				Option: MockOption{
					Timeout: types.Duration(10 * time.Second),
					Name:    "def",
				},
			}
			Expect(t, ep.Init(), Succeed())
			Expect(t, ep.Endpoint(), Equal("redis://localhost:6379/1"))
			Expect(t, ep.String(), Equal("redis://username:password@localhost:6379/1?name=abc&timeout=3s"))
			Expect(t, ep.SecurityString(), Equal("redis://localhost:6379/1?name=abc&timeout=3s"))
		})
	})

	t.Run("ModifyURLQuery", func(t *testing.T) {
		ep := types.EndpointNoOption{Address: "redis://username:password@localhost:6379/1?name=abc&timeout=3s"}
		Expect(t, ep.Init(), Succeed())

		u1 := ep.URL()
		Expect(t, u1.Query().Encode(), Equal("name=abc&timeout=3s"))
		ep.AddOption("key", "v1", "v2")
		u2 := ep.URL()
		Expect(t, u2.Query().Encode(), Equal("key=v1&key=v2&name=abc&timeout=3s"))
	})
}
