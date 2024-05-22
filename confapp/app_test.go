package confapp_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sincospro/datatypes"

	. "github.com/sincospro/conf/confapp"
)

func ExampleNewAppContext() {
	config := &struct {
		WorkerID int
		Endpoint datatypes.Endpoint
	}{
		WorkerID: 100,
	}

	if err := os.MkdirAll(filepath.Join(".", "config"), os.ModePerm); err != nil {
		return
	}

	if err := os.WriteFile(
		filepath.Join(".", "config/local.yml"),
		[]byte("TEST__WorkerID: 200\nTEST__Endpoint: postgres://username:password@hostname:5432/base?sslmode=disable"),
		os.ModePerm,
	); err != nil {
		return
	}

	app := NewAppContext(
		WithMainRoot("."),
		WithBuildMeta(Meta{
			Name:     "test",
			Feature:  "main",
			Version:  "v0.0.1",
			CommitID: "efbecda",
			Date:     "200601021504",
			Runtime:  RUNTIME_DEV,
		}),
		WithBatchRunner(
			func() {
				time.Sleep(time.Microsecond * 100)
				fmt.Println("batch runner 1")
			},
			func() {
				time.Sleep(time.Microsecond * 300)
				fmt.Println("batch runner 2")
			},
		),
		WithPreRunner(
			func() {
				time.Sleep(time.Microsecond * 100)
				fmt.Println("pre runner 1")
			},
			func() {
				time.Sleep(time.Microsecond * 300)
				fmt.Println("pre runner 2")
			},
		),
		WithMainExecutor(
			func() error {
				time.Sleep(time.Microsecond * 500)
				fmt.Println("main entry")
				return nil
			},
		),
		WithDefaultConfigGenerator(),
		WithDockerfileGenerator(),
		WithMakefileGenerator(),
	)
	app.Conf(config)

	cmd := app.Command
	buf := bytes.NewBuffer(nil)

	cmd.SetOut(buf)
	cmd.SetErr(buf)

	{
		buf.Reset()
		cmd.SetArgs([]string{"version"})
		cmd.Execute()
		fmt.Println(buf.String())
	}
	{
		buf.Reset()
		cmd.SetArgs([]string{"gen", "defaults"})
		cmd.Execute()
		content, _ := os.ReadFile(filepath.Join(".", "config/default.yml"))
		fmt.Println(string(content))
		// _ = os.RemoveAll(filepath.Join(".", "config"))
	}
	{
		buf.Reset()
		cmd.SetArgs([]string{"run"})
		cmd.Execute()
		fmt.Println(buf.String())
	}

	// Output:
	// test:main@v0.0.1#efbecda_200601021504(DEV)
	//
	// TEST__Endpoint: ""
	// TEST__WorkerID: "100"
	//
	// pre runner 1
	// pre runner 2
	// batch runner 1
	// batch runner 2
	// main entry
	// test:main@v0.0.1#efbecda_200601021504(DEV)
	// TEST__Endpoint=postgres://username:password@hostname:5432/base?sslmode=disable
	// TEST__WorkerID=200
}
