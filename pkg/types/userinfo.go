package types

import (
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/go-think/openssl"
)

// Userinfo likes url.Userinfo but embed aes decryptor if DecryptKey is set.
// and implemented SecurityString to hide password
type Userinfo struct {
	Username string
	Password Password

	// DecryptKey if not empty decrypt password by AES with ECB mode and PKCS7
	// padding. the value is the raw decrypt key, not an environment variable name.
	DecryptKey string
}

func (u *Userinfo) Init() (err error) {
	if len(u.Password) > 0 && len(u.DecryptKey) > 0 {
		var (
			cipher []byte
			plain  []byte
		)

		cipher, err = base64.StdEncoding.DecodeString(u.Password.String())
		if err != nil {
			return err
		}

		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("aes decrypt panicked: %v", r)
			}
		}()
		plain, err = openssl.AesECBDecrypt(cipher, []byte(u.DecryptKey), openssl.PKCS7_PADDING)
		if err != nil {
			return err
		}
		u.Password = Password(plain)
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
