package testdata

import (
	_ "embed"

	"github.com/xoctopus/confx/pkg/conftls"
)

var (
	//go:embed client.key
	key string
	//go:embed client.crt
	crt string
	//go:embed ca.crt
	ca string
)

func TLSConfigForTest() conftls.X509KeyPair {
	return conftls.X509KeyPair{
		Key: key,
		Crt: crt,
		CA:  ca,
	}
}
