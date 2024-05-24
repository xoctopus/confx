package confapp_test

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/sincospro/datatypes"
	"github.com/sincospro/x/misc/must"

	. "github.com/sincospro/conf/confapp"
)

func ExampleNewAppContext() {
	root := "./testdata1"
	app := NewAppContext(
		WithMainRoot(root),
		WithBuildMeta(Meta{
			Name:     "app",
			Feature:  "main",
			Version:  "v0.0.1",
			CommitID: "efbecda",
			Date:     "200601021504",
			Runtime:  RUNTIME_DEV,
		}),
		WithBatchRunner(
			func() {
				time.Sleep(time.Second * 1)
				fmt.Println("batch runner 1")
			},
			func() {
				time.Sleep(time.Second * 2)
				fmt.Println("batch runner 2")
			},
		),
		WithPreRunner(
			func() {
				time.Sleep(time.Second * 1)
				fmt.Println("pre runner 1")
			},
			func() {
				time.Sleep(time.Second * 2)
				fmt.Println("pre runner 2")
			},
		),
		WithMainExecutor(
			func() error {
				time.Sleep(time.Second * 3)
				fmt.Println("main entry")
				return nil
			},
		),
		WithDefaultConfigGenerator(),
		WithDockerfileGenerator(),
		WithMakefileGenerator(),
	)

	must.NoError(os.MkdirAll(filepath.Join(app.MainRoot(), "config"), os.ModePerm))
	must.NoError(os.WriteFile(filepath.Join(app.MainRoot(), "config/local.yml"), []byte(`
APP__CONFIG1__WorkerID: 200
APP__CONFIG1__Endpoint: postgres://username:password@hostname:5432/base?sslmode=disable
APP__CONFIG2__ServerPort: 8888
APP__CONFIG2__ClientEndpoint: http://localhost:8888/demo`), os.ModePerm))

	defer os.RemoveAll(root)

	config1 := &Config1{}
	config2 := &Config2{}
	app.Conf(config1, config2)

	cmd := app.Command
	buf := bytes.NewBuffer(nil)

	cmd.SetOut(buf)
	cmd.SetErr(buf)

	{
		fmt.Println("exec `app version`")
		buf.Reset()
		cmd.SetArgs([]string{"version"})
		must.NoError(cmd.Execute())
		fmt.Println(buf.String())
	}
	{
		fmt.Println("exec `app gen defaults`")
		buf.Reset()
		cmd.SetArgs([]string{"gen", "defaults"})
		must.NoError(cmd.Execute())
		content, _ := os.ReadFile(filepath.Join(app.MainRoot(), "config/default.yml"))
		fmt.Println(string(content))
	}
	{
		fmt.Println("exec `app run`")
		buf.Reset()
		cmd.SetArgs([]string{"run"})
		must.NoError(cmd.Execute())
		fmt.Println(buf.String())
	}

	// Output:
	// exec `app version`
	// app:main@v0.0.1#efbecda_200601021504(DEV)
	//
	// exec `app gen defaults`
	// APP__CONFIG1__Endpoint: ""
	// APP__CONFIG1__WorkerID: "0"
	// APP__CONFIG2__ClientEndpoint: http://localhost:80/demo
	// APP__CONFIG2__ServerPort: "80"
	//
	// exec `app run`
	// app:main@v0.0.1#efbecda_200601021504(DEV)
	//
	// name:     app
	// feature:  main
	// version:  v0.0.1
	// commit:   efbecda
	// date:     200601021504
	// runtime:  DEV
	//
	// APP__CONFIG1__Endpoint=postgres://username:password@hostname:5432/base?sslmode=disable
	// APP__CONFIG1__WorkerID=200
	// APP__CONFIG2__ClientEndpoint=http://localhost:8888/demo
	// APP__CONFIG2__ServerPort=8888
	//
	// pre runner 1
	// pre runner 2
	// batch runner 1
	// batch runner 2
	// main entry
}

type Config1 struct {
	WorkerID int
	Endpoint datatypes.Endpoint
}

type Config2 struct {
	ServerPort     uint16
	ClientEndpoint datatypes.Endpoint

	server *http.Server
	client *http.Client
}

func (c *Config2) SetDefault() {
	if c.ServerPort == 0 {
		c.ServerPort = 80
	}
	if c.ClientEndpoint.IsZero() {
		must.NoError(c.ClientEndpoint.UnmarshalText([]byte("http://localhost:80/demo")))
	}
}
func (c *Config2) Init() {
	c.server = &http.Server{
		Addr: fmt.Sprintf(":%d", c.ServerPort),
	}
}

type MustInitFailed struct{}

func (i *MustInitFailed) Init() error { return errors.New("must fail") }

func ExampleInitFailed() {
	os.Setenv("APP__SomeKey", "")

	app := NewAppContext(
		WithBuildMeta(Meta{Name: "app"}),
	)

	defer func() {
		fmt.Println(recover())
	}()

	app.Conf(&struct {
		SomeKey MustInitFailed
	}{})

	// Output:
	// failed to init [group:APP] [field:SomeKey]: must fail
}
