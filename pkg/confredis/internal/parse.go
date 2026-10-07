package internal

import "net/url"

// officialQueries lists query keys understood by go-redis ParseURL (setupConnParams).
// Sync when bumping github.com/redis/go-redis/v9 (currently v9.23.0).
var officialQueries = map[string]struct{}{
	"client_name":                {},
	"conn_max_idle_time":         {},
	"conn_max_lifetime":          {},
	"conn_max_lifetime_jitter":   {},
	"db":                         {},
	"dial_timeout":               {},
	"idle_timeout":               {},
	"max_active_conns":           {},
	"max_concurrent_dials":       {},
	"max_conn_age":               {},
	"max_idle_conns":             {},
	"max_retries":                {},
	"max_retry_backoff":          {},
	"min_idle_conns":             {},
	"min_retry_backoff":          {},
	"pipeline_pool_size":         {},
	"pipeline_read_buffer_size":  {},
	"pipeline_write_buffer_size": {},
	"pool_fifo":                  {},
	"pool_size":                  {},
	"pool_timeout":               {},
	"protocol":                   {},
	"read_timeout":               {},
	"skip_verify":                {},
	"write_timeout":              {},
}

// externalQueries lists URL query keys for confredis.Option (textx.UnmarshalURL naming).
// Keep in sync with Option struct fields; SentinelAuth uses url:"-" and is omitted.
var externalQueries = map[string]struct{}{
	"prefix":            {},
	"db":                {},
	"enableTls":         {},
	"connectionTimeout": {},
	"operationTimeout":  {},
	"bufferSizeKb":      {},
	"poolSize":          {},
	"maxIdleConnection": {},
	"maxIdleTime":       {},
	"masterName":        {},
	"clusterMode":       {},
}

// ParseURL splits Address query into go-redis keys and confx Option keys.
// db is passed to both. Other query keys are ignored.
func ParseURL(raw string) (_ string, external url.Values, err error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", nil, err
	}

	official := url.Values{} // go-redis official supported queries
	external = url.Values{}
	for k, vs := range u.Query() {
		_, confx := externalQueries[k]
		_, redis := officialQueries[k]
		switch {
		case confx && redis:
			official[k] = vs
			external[k] = vs
		case confx:
			external[k] = vs
		case redis:
			official[k] = vs
		default:
			// ignore query keys neither go-redis nor confx Option understand
		}
	}

	u.RawQuery = official.Encode()
	return u.String(), external, nil
}
