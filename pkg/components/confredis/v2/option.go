package confredis

import (
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/textx"

	"github.com/xoctopus/confx/pkg/types"
)

type Option struct {
	// Prefix
	Prefix string
	// Addresses cluster address
	Addresses []string `url:"-"`
	// DB single-node database to be selected once connected
	DB int `url:""`
	// ConnectionTimeout connection timeout
	ConnectionTimeout types.Duration `url:",default=100ms"`
	// OperationTimeout read/write timeout
	OperationTimeout types.Duration `url:",default=100ms"`

	// BufferSizeKB is the size of the bufio.Reader buffer for each connection.
	// default: 128KiB
	BufferSizeKB int `url:",default=128"`

	// PoolSize controls redis request concurrency. 10*GOMAXPROCS is recommended.
	// this option affects MaxActiveConns, which will be set to 4 times of PoolSize
	// for balancing connection stability and concurrent performance.
	//edis connection stability and concurrent performance.
	PoolSize          int            `url:",default=20"`
	MaxIdleConnection int            `url:",default=10"`
	MaxIdleTime       types.Duration `url:",default=1h"`

	// SentinelAuth auth info for sentinel if enabled
	SentinelAuth types.Userinfo `url:"-"`
	// MasterName is sentinel master name
	MasterName string

	ClusterMode bool
}

func (o *Option) SetDefault() {
	must.NoError(textx.UnmarshalURL(url.Values{}, o))
}

func (o Option) ClientOption() *redis.UniversalOptions {
	return &redis.UniversalOptions{
		DB:         o.DB,
		ClientName: o.Prefix,

		DialTimeout:           time.Duration(o.ConnectionTimeout),
		ReadTimeout:           time.Duration(o.OperationTimeout),
		WriteTimeout:          time.Duration(o.OperationTimeout),
		ContextTimeoutEnabled: true,

		ReadBufferSize:  o.BufferSizeKB * 1024,
		WriteBufferSize: o.BufferSizeKB * 1024,

		PoolSize:        o.PoolSize,
		PoolTimeout:     time.Duration(o.ConnectionTimeout),
		MinIdleConns:    max(3, o.MaxIdleConnection/2),
		MaxIdleConns:    o.MaxIdleConnection,
		MaxActiveConns:  4 * o.PoolSize,
		ConnMaxIdleTime: time.Duration(o.MaxIdleTime),

		IdentitySuffix: o.Prefix + "_v9",

		SentinelUsername: o.SentinelAuth.Username,
		SentinelPassword: o.SentinelAuth.Password.String(),
		IsClusterMode:    o.ClusterMode,
	}
}
