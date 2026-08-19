package types_test

import (
	"encoding/base64"
	"net/url"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/types"
)

func TestUserinfo(t *testing.T) {
	u := &types.Userinfo{
		Username:   "username",
		Password:   "LelzsnHN2xnJd/MB+JGIXWqd8pJPhPYfuRfDbrCsZE8=",
		DecryptKey: "9f67229a84e2229ee9a834c151d068f5",
	}

	Expect(t, u.IsZero(), BeFalse())
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
		u = &types.Userinfo{
			Username:   "user",
			Password:   "abc$%^",
			DecryptKey: "9f67229a84e2229ee9a834c151d068f5",
		}
		Expect(t, u.Init(), Failed())
	})

	t.Run("FailedAesDecode", func(t *testing.T) {
		u = &types.Userinfo{
			Username:   "username",
			Password:   "LelzsnHN2xnJd/MB+JGIXWqd8pJPhPYfuRfDbrCsZE8=",
			DecryptKey: "0123456789abcde",
		}
		Expect(t, u.Init(), Failed())

		u.DecryptKey = "def"
		u.Password = types.Password(base64.StdEncoding.EncodeToString([]byte("abc")))
		Expect(t, u.Init(), ErrorContains("aes decrypt panicked"))
	})
}
