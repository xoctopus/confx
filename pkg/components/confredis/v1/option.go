package confredis

import "github.com/xoctopus/confx/pkg/types"

type RedisOptions struct {
	Prefix        string         `url:""`
	ConnTimeout   types.Duration `url:",default=10s"`
	WriteTimeout  types.Duration `url:",default=10s"`
	ReadTimeout   types.Duration `url:",default=10s"`
	KeepAlive     types.Duration `url:",default=3600s"`
	MaxActive     int            `url:",default=100"`
	MaxIdle       int            `url:",default=100"`
	ClientName    string         `url:""`
	SkipTLSVerify bool           `url:",default=true"`
	Wait          bool           `url:",default=true"`
}
