package option

import (
	"crypto/tls"

	"github.com/go-sql-driver/mysql"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/textx"

	"github.com/xoctopus/confx/pkg/types"
)

type MySQL struct {
	Charset           string
	Collation         string
	InterpolateParams bool
	Loc               string         `url:",defualt=Asia/Shanghai"`
	MultiStatements   bool           `url:",default=true"`
	ParseTime         bool           `url:",default=true"`
	Timeout           types.Duration `url:",default=2s"`
	ReadTimeout       types.Duration `url:",default=30s"`
	WriteTimeout      types.Duration `url:",default=10s"`
	TLS               string         `url:",default=false"`
}

func (o *MySQL) SetDefault() {
	_ = must.NoErrorV(textx.SetDefault(o))
}

func (o *MySQL) WithTLS(c *tls.Config) error {
	return mysql.RegisterTLSConfig(o.TLS, c)
}
