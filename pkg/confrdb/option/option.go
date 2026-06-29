package option

import "crypto/tls"

type TLSConfigPatcher interface {
	WithTLS(*tls.Config) error
}

type Overwriter interface {
	Overwrite()
}
