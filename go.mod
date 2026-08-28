module github.com/xoctopus/confx

go 1.27.0

tool (
	github.com/xoctopus/confx/internal/cmd/gen
	github.com/xoctopus/confx/internal/cmd/skill-install
)

// message queue
//
// utils
require (
	// mq:pulsar
	github.com/apache/pulsar-client-go v0.21.0
	github.com/fatih/color v1.19.0
	// rdb:mysql
	github.com/go-sql-driver/mysql v1.10.0
	github.com/go-think/openssl v1.22.0
	// authorization
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/gorilla/websocket v1.5.3
	// rdb:postgres
	github.com/jackc/pgx/v5 v5.10.0
	github.com/oklog/ulid/v2 v2.1.2
	// otel
	github.com/prometheus/client_golang v1.24.1
	// mq:rabbit
	github.com/rabbitmq/amqp091-go v1.13.0
	// kv-storage:redis
	github.com/redis/go-redis/v9 v9.22.0
	// mq:kafka
	github.com/segmentio/kafka-go v0.4.51
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
	// mq:rabbit
	github.com/wagslane/go-rabbitmq v0.16.1
	// +skill:concx
	github.com/xoctopus/concx v0.2.0
	// +skill:genx
	github.com/xoctopus/genx v0.3.8
	github.com/xoctopus/httpx v0.0.2
	// +skill:logx
	github.com/xoctopus/logx v0.3.8
	github.com/xoctopus/sfid v0.1.0
	// +skill:sqlx
	github.com/xoctopus/sqlx v0.3.8
	// +skill:testx
	github.com/xoctopus/x v0.5.8
	// otel
	go.opentelemetry.io/otel v1.46.0
	// otel
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.44.0
	// otel
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.44.0
	// otel
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.44.0
	// otel @unstable @easy-breaking pin this section
	go.opentelemetry.io/otel/exporters/prometheus v0.66.0
	// otel @unstable @easy-breaking pin this section
	go.opentelemetry.io/otel/log v0.21.0
	// otel
	go.opentelemetry.io/otel/metric v1.46.0
	// otel
	go.opentelemetry.io/otel/sdk v1.45.0
	// otel @unstable @easy-breaking pin this section
	go.opentelemetry.io/otel/sdk/log v0.21.0
	// otel
	go.opentelemetry.io/otel/sdk/metric v1.45.0
	// otel
	go.opentelemetry.io/otel/trace v1.46.0
	// zap logger for stashing
	go.uber.org/zap v1.28.0
	golang.org/x/sync v0.22.0
	gopkg.in/yaml.v3 v3.0.1
	// rdb:sqlite
	modernc.org/sqlite v1.55.0
)

// indirect
require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/AthenZ/athenz v1.12.13 // indirect
	github.com/DataDog/zstd v1.5.0 // indirect
	github.com/RoaringBitmap/roaring/v2 v2.8.0 // indirect
	github.com/ardielle/ardielle-go v1.5.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.20.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.9.1 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	github.com/hamba/avro/v2 v2.29.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.19.1 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/onsi/gomega v1.38.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.27 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.70.1 // indirect
	github.com/prometheus/otlptranslator v1.0.0 // indirect
	github.com/prometheus/procfs v0.21.1 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xoctopus/pkgx v0.4.4 // indirect
	github.com/xoctopus/typx v0.4.7 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.63.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	golang.org/x/mod v0.40.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/tools v0.49.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260630182238-925bb5da69e7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260630182238-925bb5da69e7 // indirect
	google.golang.org/grpc v1.82.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/apimachinery v0.37.0 // indirect
	k8s.io/client-go v0.32.3 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/kube-openapi v0.0.0-20260721132016-d427ff9ee9ad // indirect
	k8s.io/utils v0.0.0-20260626114624-be93311217bd // indirect
	modernc.org/libc v1.74.1 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.4.2 // indirect
)
