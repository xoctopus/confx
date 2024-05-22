package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/sincospro/datatypes"

	"github.com/sincospro/conf/confapp"
	"github.com/sincospro/conf/confapp/example/pkg/modules/module1"
	"github.com/sincospro/conf/confapp/example/pkg/modules/module2"
	"github.com/sincospro/conf/confapp/example/pkg/modules/module3"
)

var (
	Name     = "example"
	Feature  string
	Version  string
	CommitID string
	Date     string

	app    *confapp.AppCtx
	config = &struct {
		WorkerID uint64
		Endpoint datatypes.Endpoint
	}{
		WorkerID: 100,
	}
)

func init() {
	app = confapp.NewAppContext(
		confapp.WithBuildMeta(confapp.Meta{
			Name:     Name,
			Feature:  Feature,
			Version:  Version,
			CommitID: CommitID,
			Date:     Date,
		}),
		confapp.WithMainRoot("."),
		confapp.WithDefaultConfigGenerator(),
		confapp.WithMainExecutor(Main),
		confapp.WithPreRunner(
			module1.InitRunner(context.Background()),
		),
		confapp.WithBatchRunner(
			module2.InitRunner(context.Background()),
			module3.InitRunner(context.Background()),
		),
	)

	app.Conf(config)
}

func Main() error {
	log.Printf("app: %s", app.Version())
	log.Printf("WorkerID: %v", config.WorkerID)
	log.Printf("Endpoint: %s", config.Endpoint)
	time.Sleep(2 * time.Second)
	return nil
}

func main() {
	if err := app.Command.Execute(); err != nil {
		app.PrintErrln(err)
		os.Exit(-1)
	}
}
