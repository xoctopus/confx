# confx

<p align="left">
  <a href="https://github.com/xoctopus/confx"><img src="https://img.shields.io/github/go-mod/go-version/xoctopus/confx?logo=go&style=flat-square&logoColor=white" alt="Go Version"></a>
  <a href="https://github.com/xoctopus/confx/actions/workflows/lint.yml"><img src="https://img.shields.io/github/actions/workflow/status/xoctopus/confx/lint.yml?label=golangci-lint&logo=go&style=flat-square" alt="golangci-lint"></a>
  <a href="https://codecov.io/gh/xoctopus/confx"><img src="https://img.shields.io/codecov/c/github/xoctopus/confx?token=XR9205WT7D&logo=codecov&style=flat-square" alt="codecov"></a>
  <a href="https://pkg.go.dev/github.com/xoctopus/confx"><img src="https://pkg.go.dev/badge/github.com/xoctopus/confx.svg" alt="Go Reference"></a>
  <a href="https://github.com/xoctopus/confx/blob/main/LICENSE"><img src="https://img.shields.io/github/license/xoctopus/confx?style=flat-square" alt="License"></a>
</p>

---

confx is the application layer of [xoctopus](https://github.com/xoctopus): it
turns a plain Go struct into a configured, observable, and gracefully stoppable
service.

Declare the configuration and infrastructure a service needs as struct fields.
confx binds them to environment variables, generates configuration templates,
initializes each component in order, and closes them on shutdown.

- **Configuration as types**: struct fields map to environment variables, with
  defaults taken from your Go values.
- **Unified component lifecycle**: components share one initialization and
  shutdown contract.
- **Fail fast**: a missing address or an unreachable dependency stops the service
  at startup, not on the first request.
- **Ready-made infrastructure**: Redis, relational databases, Pulsar, RabbitMQ,
  WebSocket, JWT, OpenTelemetry, and more.

> [!NOTE]
> confx is pre-1.0, so APIs may change between minor versions.

## Install

Requires Go 1.27.3 or later.

```sh
go get github.com/xoctopus/confx
```

## Quick start

```go
package main

import (
	"context"
	"log"

	"github.com/xoctopus/confx/pkg/appx"
	"github.com/xoctopus/confx/pkg/types"
)

var (
	app *appx.AppCtx
	ctx = context.Background()

	// Field values are the defaults.
	config = &struct {
		WorkerID uint64
		Endpoint types.EndpointNoOption
	}{WorkerID: 100}
)

func init() {
	app = appx.NewAppContext(
		Main,
		appx.WithMeta(appx.Meta{
			Name:    "example",
			Version: "v0.1.0",
			Runtime: appx.GetRuntime(),
		}),
		appx.WithRoot("."),
	)
	ctx = app.Conf(ctx, config)
}

func Main() error {
	defer func() { _ = app.Close(ctx) }()

	log.Printf("worker=%d endpoint=%q", config.WorkerID, config.Endpoint.Address)
	return nil
}

func main() {
	if err := app.Execute(); err != nil {
		log.Fatal(err)
	}
}
```

Variable names combine the application name and the field path:

```sh
go run . run
# worker=100 endpoint=""

EXAMPLE__WorkerID=200 EXAMPLE__Endpoint_Address=redis://127.0.0.1:6379 go run . run
# worker=200 endpoint="redis://127.0.0.1:6379"

go run . version
# example:@v0.1.0#(PROD)[commit=|build=]
```

Each run also writes the available variables and their defaults to
`config/default.yml` and `config/.env`. Values can be overridden locally in
`config/local.yml`.

### Add a component

A component is just another field. This one connects to Redis during `app.Conf`:

```go
config = &struct {
	Redis confredis.Endpoint
}{}
```

```sh
EXAMPLE__Redis_Address=redis://127.0.0.1:6379 go run . run
```

The service panics at startup if the address is empty or Redis is unreachable.

## Application lifecycle

```text
app.Conf     load config -> write defaults -> PreInit -> initialize components
app.Execute  PreRun -> Serve -> Main
app.Close    close components -> WithClose hooks
```

`Main` decides whether to wait or exit, and calls `app.Close`. One-shot commands,
such as migrations, can be added with `app.AddCommand`.

## Packages

### Application

| Package                 | Purpose                                             |
|-------------------------|-----------------------------------------------------|
| [`appx`](pkg/appx)      | Application context, startup hooks, and shutdown    |
| [`cmdx`](pkg/cmdx)      | Cobra commands and flags from tagged structs        |
| [`envx`](pkg/envx)      | Encoding and decoding structs as environment values |
| [`types`](pkg/types)    | Endpoint, lifecycle, liveness, KV, and MQ contracts |

### Components

| Package                       | Purpose                                         |
|-------------------------------|-------------------------------------------------|
| [`confredis`](pkg/confredis)  | Redis client                                    |
| [`confrdb`](pkg/confrdb)      | MySQL, PostgreSQL, and SQLite through sqlx      |
| [`confpulsar`](pkg/confpulsar)| Pulsar producer and consumer endpoints          |
| [`confrabbit`](pkg/confrabbit)| RabbitMQ producer and consumer endpoints        |
| [`confws`](pkg/confws)        | WebSocket server and client management          |
| [`confxxl`](pkg/confxxl)      | XXL-JOB executor registration and job dispatch  |
| [`confjwt`](pkg/confjwt)      | JWT signing, validation, and context injection  |
| [`conftls`](pkg/conftls)      | TLS and X.509 certificates                      |
| [`conflogx`](pkg/conflogx)    | Log level, format, backend, and file output     |
| [`confotel`](pkg/confotel)    | OpenTelemetry traces, metrics, and logs         |

`confkafka`, `confbranca`, and `confethcli` are incomplete and not ready for use.

## Examples

- [`example/appx`](example/appx): an application with configuration, startup
  hooks, and detached modules
- [`example/cmdx`](example/cmdx): a CLI defined with tagged structs

## Development

```sh
make test   # start dependencies from hack/ and run tests
make lint
make cover
```

## License

[MIT](LICENSE)
