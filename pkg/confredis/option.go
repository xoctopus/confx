package confredis

import (
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/slicex"
	"github.com/xoctopus/x/textx"

	"github.com/xoctopus/confx/pkg/types"
)

type Option struct {
	// Prefix
	Prefix string
	// DB single-node database to be selected once connected
	DB int `url:""`
	// EnableTLS forces TLS even when Address scheme is redis://.
	// rediss:// always enables TLS regardless of this flag.
	EnableTLS bool `url:""`
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
	must.NoErrorV(textx.SetDefault(o))
}

// ClientOption builds redis.UniversalOptions from main URL and optional seed addresses.
func (o Option) ClientOption(main string, others ...string) *redis.UniversalOptions {
	must.BeTrueF(len(main) > 0, "empty main address")

	// main is parsed by redis.ParseURL.
	// IMPORTANT: go-redis rejects any URL query key they dont know
	// TODO: raise this with go-redis (ignore unknown query vs hard fail).
	base := must.NoErrorV(redis.ParseURL(main))

	seeds := []string{base.Addr}
	for _, s := range others {
		seeds = append(seeds, must.NoErrorV(HostPort(s)))
	}

	db := base.DB
	if o.DB != 0 {
		db = o.DB
	}
	if o.ClusterMode {
		db = 0
	}

	return &redis.UniversalOptions{
		Addrs:      slicex.Unique(seeds),
		DB:         db,
		ClientName: o.Prefix,
		Username:   base.Username,
		Password:   base.Password,
		Protocol:   base.Protocol,

		MaxRetries:      base.MaxRetries,
		MinRetryBackoff: base.MinRetryBackoff,
		MaxRetryBackoff: base.MaxRetryBackoff,

		DialTimeout:           time.Duration(o.ConnectionTimeout),
		ReadTimeout:           time.Duration(o.OperationTimeout),
		WriteTimeout:          time.Duration(o.OperationTimeout),
		ContextTimeoutEnabled: true,
		PoolTimeout:           time.Duration(o.ConnectionTimeout),

		ReadBufferSize:  o.BufferSizeKB * 1024,
		WriteBufferSize: o.BufferSizeKB * 1024,

		PoolFIFO:              base.PoolFIFO,
		PoolSize:              o.PoolSize,
		MinIdleConns:          max(3, o.MaxIdleConnection/2),
		MaxIdleConns:          o.MaxIdleConnection,
		MaxActiveConns:        4 * o.PoolSize,
		ConnMaxIdleTime:       time.Duration(o.MaxIdleTime),
		ConnMaxLifetime:       base.ConnMaxLifetime,
		ConnMaxLifetimeJitter: base.ConnMaxLifetimeJitter,

		IdentitySuffix: o.Prefix + "_v9",

		MasterName:       o.MasterName,
		SentinelUsername: o.SentinelAuth.Username,
		SentinelPassword: o.SentinelAuth.Password.String(),
		IsClusterMode:    o.ClusterMode,
	}
}
