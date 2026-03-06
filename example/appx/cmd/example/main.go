package main

import (
	"context"
	"log"
	"os"
	"time"

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

	app    *appx.AppCtx
	config = &struct {
		WorkerID uint64
		Endpoint types.EndpointNoOption
	}{
		WorkerID: 100,
	}
)

func init() {
	// try to override local.yml
	_ = os.Setenv("EXAMPLE__WorkerID", "111")

	app = appx.NewAppContext(
		appx.WithBuildMeta(appx.Meta{
			Name:     Name,
			Feature:  Feature,
			Version:  Version,
			CommitID: CommitID,
			Date:     Date,
		}),
		appx.WithMainRoot("."),
		appx.WithMainExecutor(Main),
		appx.WithPreRunner(
			module1.InitRunner(context.Background()),
			module2.InitRunner(context.Background()),
		),
		appx.WithBatchRunner(
			module3.InitRunner(context.Background()),
		),
	)

	app.Conf(context.Background(), config)
}

func Main() error {
	log.Printf("app: %s", app.Version())
	log.Printf("WorkerID: %v", config.WorkerID)
	log.Printf("Endpoint: %s", config.Endpoint.Address)
	time.Sleep(2 * time.Second)
	return nil
}

func main() {
	if err := app.Command.Execute(); err != nil {
		app.PrintErrln(err)
		os.Exit(-1)
	}
}
