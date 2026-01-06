package types

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"

	"github.com/go-think/openssl"
)

// Userinfo likes url.Userinfo but embed aes decryptor if DecryptKeyEnv is set.
// and implemented SecurityString to hide password
type Userinfo struct {
	Username string
	Password Password

	// DecryptKeyEnv if not empty read key from env and decrypt password by AEC
	// with ECB mode and PKCS7 padding. default env key is PASSWORD_DEC_KEY
	DecryptKeyEnv string
}

func (u *Userinfo) SetDefault() {
	if u.DecryptKeyEnv == "" {
		u.DecryptKeyEnv = "PASSWORD_DEC_KEY"
	}
}

func (u *Userinfo) Init() (err error) {
	if len(u.Password) > 0 {
		var (
			cipher []byte
			plain  []byte
			key    []byte
		)

		key = []byte(os.Getenv(u.DecryptKeyEnv))
		if len(key) > 0 {
			cipher, err = base64.StdEncoding.DecodeString(u.Password.String())
			if err != nil {
				return err
			}

			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("aes decrypt panicked: %v", r)
				}
			}()
			plain, err = openssl.AesECBDecrypt(cipher, key, openssl.PKCS7_PADDING)
			if err != nil {
				return err
			}
			u.Password = Password(plain)
		}
	}
	return nil
}

func (u *Userinfo) IsZero() bool {
	return u.Username == "" && u.Password == ""
}

func (u Userinfo) String() string {
	if u.IsZero() {
		return ""
	}
	if len(u.Password) == 0 {
		return u.Username
	}
	return u.Username + ":" + u.Password.String()
}

func (u Userinfo) SecurityString() string {
	if u.IsZero() {
		return ""
	}
	if len(u.Password) == 0 {
		return u.Username
	}
	return u.Username + ":" + u.Password.SecurityString()
}

func (u Userinfo) Userinfo() *url.Userinfo {
	if u.Password == "" {
		return url.User(u.Username)
	}
	return url.UserPassword(u.Username, u.Password.String())
}
