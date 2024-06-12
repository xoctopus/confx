package conftls

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/pkg/errors"
)

var DefaultTLSConfig = &tls.Config{
	ClientAuth:         tls.NoClientCert,
	ClientCAs:          nil,
	InsecureSkipVerify: true,
}

type X509KeyPair struct {
	Key string
	Crt string
	CA  string

	config *tls.Config
}

func (c *X509KeyPair) IsZero() bool {
	return c.Key == "" || c.CA == "" || c.Crt == ""
}

func (c *X509KeyPair) read() (key, crt, ca []byte) {
	var err error

	key, err = os.ReadFile(c.Key)
	if err != nil {
		goto ReturnValue
	}
	crt, err = os.ReadFile(c.Crt)
	if err != nil {
		goto ReturnValue
	}
	ca, err = os.ReadFile(c.CA)
	if err != nil {
		goto ReturnValue
	}
	return key, crt, ca

ReturnValue:
	return []byte(c.Key), []byte(c.Crt), []byte(c.CA)
}

func (c *X509KeyPair) Init() error {
	if c == nil || c.IsZero() {
		return nil
	}

	key, crt, ca := c.read()

	cert, err := tls.X509KeyPair(crt, key)
	if err != nil {
		return err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(ca) {
		return errors.Errorf("failed to append cert")
	}
	c.config = &tls.Config{
		RootCAs:            pool,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
	return nil
}

func (c *X509KeyPair) Config() *tls.Config {
	if c == nil || c.config == nil || c.IsZero() {
		return DefaultTLSConfig
	}
	return c.config
}
