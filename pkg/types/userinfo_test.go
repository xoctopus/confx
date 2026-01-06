package types_test

import (
	"encoding/base64"
	"net/url"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/types"
)

func TestUserinfo(t *testing.T) {
	t.Setenv("PASSWORD_DEC_KEY", "9f67229a84e2229ee9a834c151d068f5")

	u := &types.Userinfo{
		Username: "username",
		Password: "LelzsnHN2xnJd/MB+JGIXWqd8pJPhPYfuRfDbrCsZE8=",
	}
	u.SetDefault()

	Expect(t, u.IsZero(), BeFalse())
	Expect(t, u.DecryptKeyEnv, Equal("PASSWORD_DEC_KEY"))
	Expect(t, u.Init(), Succeed())
	Expect(t, u.Password.String(), Equal("rhdsicyjzbwbtdwnxcei"))
	Expect(t, u.String(), Equal("username:rhdsicyjzbwbtdwnxcei"))
	Expect(t, u.SecurityString(), Equal("username:"+types.MaskedPassword))

	Expect(t, (types.Userinfo{}).String(), Equal(""))
	Expect(t, (types.Userinfo{}).SecurityString(), Equal(""))
	Expect(t, (types.Userinfo{Username: "user"}).String(), Equal("user"))
	Expect(t, (types.Userinfo{Username: "User"}).SecurityString(), Equal("User"))
	Expect(t, (types.Userinfo{Username: "user"}).Userinfo(), Equal(url.User("user")))
	Expect(t, (types.Userinfo{Username: "user", Password: "pass"}).String(), Equal("user:pass"))
	Expect(t, (types.Userinfo{Username: "User", Password: "pass"}).SecurityString(), Equal("User:--------"))
	Expect(t, (types.Userinfo{Username: "User", Password: "pass"}).Userinfo(), Equal(url.UserPassword("User", "pass")))

	t.Run("InvalidBase64Password", func(t *testing.T) {
		t.Setenv("PASSWORD_DEC_KEY", "9f67229a84e2229ee9a834c151d068f5")
		u = &types.Userinfo{Username: "user", Password: "abc$%^"}
		u.SetDefault()
		Expect(t, u.Init(), Failed())
	})

	t.Run("FailedAesDecode", func(t *testing.T) {
		t.Setenv("PASSWORD_DEC_KEY_INVALID", "0123456789abcde")
		u = &types.Userinfo{
			Username:      "username",
			Password:      "LelzsnHN2xnJd/MB+JGIXWqd8pJPhPYfuRfDbrCsZE8=",
			DecryptKeyEnv: "PASSWORD_DEC_KEY_INVALID",
		}
		Expect(t, u.Init(), Failed())

		t.Setenv("PASSWORD_DEC_KEY_INVALID_2", "def")
		u.DecryptKeyEnv = "PASSWORD_DEC_KEY_INVALID_2"
		u.Password = types.Password(base64.StdEncoding.EncodeToString([]byte("abc")))
		Expect(t, u.Init(), ErrorContains("aes decrypt panicked"))
	})
}
