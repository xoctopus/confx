package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/xoctopus/logx"

	"github.com/xoctopus/confx/example/appx/pkg/modules/module1"
	"github.com/xoctopus/confx/example/appx/pkg/modules/module2"
	"github.com/xoctopus/confx/example/appx/pkg/modules/module3"
	"github.com/xoctopus/confx/pkg/appx"
	"github.com/xoctopus/confx/pkg/types"
)

var (
	Name     = "example"
	Feature  string
	Version  string
	CommitID string
	Date     string

	meta appx.Meta

	// singleton app
	app *appx.AppCtx
	// global context
	ctx, cancel = context.WithCancel(context.Background())

	// define global configurations and set default
	config = &struct {
		WorkerID  uint64
		Endpoint  types.EndpointNoOption
		LogLevel  logx.LogLevel
		LogFormat logx.LogFormat
	}{
		WorkerID:  100,
		LogLevel:  logx.LogLevelDebug,
		LogFormat: logx.LogFormatJSON,
	}
)

// force override configurations from env or if needed.
func init() {
	_ = os.Setenv("EXAMPLE__WorkerID", "111")
}

// initialize appx.Meta [required]
func init() {
	meta = appx.Meta{
		Name:     Name,
		Feature:  Feature,
		Version:  Version,
		CommitID: CommitID,
		Date:     Date,
	}
}

// initialize appx.AppCtx and [required]
func init() {
	app = appx.NewAppContext(
		Main,
		appx.WithMeta(meta),
		appx.WithRoot("."),
		appx.WithClose(),
		appx.WithPreRun(
			// initialize global configurations.
			func() {
				fmt.Println("global configurations are initialized")
			},
			// inject global context
			func() {
				fmt.Println("global contexts are injected")
			},
		),
		appx.WithServe(
			// start http server
			func() {
				fmt.Println("http server listening on :80")
			},
			// other runners can be detached
			module1.InitRunner(context.Background()),
			module2.InitRunner(context.Background()),
			module3.InitRunner(context.Background()),
		),
	)

	app.Conf(ctx, config)
}

// Main app main entry
func Main() error {
	defer func() {
		if err := app.Close(ctx); err != nil {
			log.Println(err)
		}
		cancel()
	}()

	log.Printf("app: %s", app.Version())
	log.Printf("WorkerID: %v", config.WorkerID)
	log.Printf("Endpoint: %s", config.Endpoint.Address)
	time.Sleep(2 * time.Second)
	return nil
}

func main() {
	if err := app.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
