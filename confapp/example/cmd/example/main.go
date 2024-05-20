package main

import (
	"log"
	"os"

	"github.com/sincospro/datatypes"

	"github.com/sincospro/conf/confapp"
)

var (
	Name     = "example"
	Feature  string
	Version  string
	CommitID string
	Date     string

	app *confapp.AppCtx
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
		confapp.WithWorkDir("."),
	)
}

func main() {
	log.Printf("==> app version")
	log.Println(app.Meta.String())

	config := &struct {
		WorkerID uint64
		Endpoint datatypes.Endpoint
	}{
		WorkerID: 100,
	}

	log.Println("==> default config:")
	log.Printf("WorkerID: %v", config.WorkerID)
	log.Printf("Endpoint: %s", config.Endpoint)

	content, err := os.ReadFile("config/local.yml")
	if err != nil {
		panic(err)
	}
	log.Println("==>local config:")
	log.Printf("\n%s", string(content))

	if err = app.Conf(config); err != nil {
		panic(err)
	}

	log.Println("==>after local config loaded: ")
	log.Printf("WorkerID: %v", config.WorkerID)
	log.Printf("Endpoint: %s", config.Endpoint)
}
