package appx_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xoctopus/x/misc/must"
	. "github.com/xoctopus/x/testx"

	. "github.com/xoctopus/confx/pkg/appx"
	"github.com/xoctopus/confx/pkg/envx"
	"github.com/xoctopus/confx/pkg/types"
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
	app.Conf(context.Background(), config1, config2)

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
	// APP__CONFIG1__Endpoint=postgres://username:--------@hostname:5432/base?sslmode=disable
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
	Endpoint types.Endpoint
}

type Config2 struct {
	ServerPort     uint16
	ClientEndpoint types.Endpoint

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

func Example_init_failed() {
	_ = os.Setenv("APP__SomeKey", "")

	root := "./testdata2"
	app := NewAppContext(
		WithBuildMeta(Meta{Name: "app"}),
		WithMainRoot(root),
	)

	defer os.RemoveAll(root)

	defer func() {
		fmt.Println(recover())
	}()

	app.Conf(context.Background(), &struct {
		SomeKey MustInitFailed
	}{})

	// Output:
	// failed to init [group:APP] [field:SomeKey]: must fail
}

type MapConfig struct {
	Number *big.Int
	String string
}

type config struct {
	Map map[string]*MapConfig
}

func TestAppCtx_Conf(t *testing.T) {
	t.Run("ConfSlice", func(t *testing.T) {
		t.Run("HasEnvVars", func(t *testing.T) {
			t.Setenv("TEST__IntSlice_0", "10")
			t.Setenv("TEST__IntSlice_2", "12")
			t.Setenv("TEST__IntSlice_4", "14")
			t.Run("DefaultEmpty", func(t *testing.T) {
				app := NewAppContext(WithBuildMeta(Meta{Name: "TEST"}))
				defer os.RemoveAll(filepath.Join(app.MainRoot(), "config"))
				v := &struct{ IntSlice []int }{}
				app.Conf(context.Background(), v)
				Expect(t, v.IntSlice, Equal([]int{10, 0, 12, 0, 14}))
			})
			t.Run("HasDefaultValue", func(t *testing.T) {
				app := NewAppContext(WithBuildMeta(Meta{Name: "TEST"}))
				defer os.RemoveAll(filepath.Join(app.MainRoot(), "config"))
				v := &struct{ IntSlice []int }{IntSlice: []int{1, 2, 3}}
				app.Conf(context.Background(), v)
				Expect(t, v.IntSlice, Equal([]int{10, 0, 12, 0, 14}))
			})
		})
		t.Run("NoEnvVars", func(t *testing.T) {
			t.Run("DefaultEmpty", func(t *testing.T) {
				app := NewAppContext(WithBuildMeta(Meta{Name: "TEST"}))
				defer os.RemoveAll(filepath.Join(app.MainRoot(), "config"))
				v := &struct{ IntSlice []int }{}
				app.Conf(context.Background(), v)
				Expect(t, v.IntSlice, BeNil[[]int]())
			})
			t.Run("HasDefaultValue", func(t *testing.T) {
				app := NewAppContext(WithBuildMeta(Meta{Name: "TEST"}))
				defer os.RemoveAll(filepath.Join(app.MainRoot(), "config"))
				v := &struct{ IntSlice []int }{IntSlice: []int{1, 2, 3}}
				app.Conf(context.Background(), v)
				Expect(t, v.IntSlice, Equal([]int{1, 2, 3}))
			})
		})
	})
	t.Run("ConfArray", func(t *testing.T) {
		t.Run("HasEnvVars", func(t *testing.T) {
			t.Setenv("TEST__IntArray_0", "10")
			t.Setenv("TEST__IntArray_2", "12")
			t.Setenv("TEST__IntArray_4", "14")
			t.Run("DefaultEmpty", func(t *testing.T) {
				app := NewAppContext(WithBuildMeta(Meta{Name: "TEST"}))
				defer os.RemoveAll(filepath.Join(app.MainRoot(), "config"))
				v := &struct{ IntArray [3]int }{}
				app.Conf(context.Background(), v)
				Expect(t, v.IntArray, Equal([3]int{10, 0, 12}))
			})
			t.Run("HasDefaultValue", func(t *testing.T) {
				app := NewAppContext(WithBuildMeta(Meta{Name: "TEST"}))
				defer os.RemoveAll(filepath.Join(app.MainRoot(), "config"))
				v := &struct{ IntArray [3]int }{IntArray: [3]int{1, 2, 3}}
				app.Conf(context.Background(), v)
				Expect(t, v.IntArray, Equal([3]int{10, 2, 12}))
			})
		})
		t.Run("NoEnvVars", func(t *testing.T) {
			t.Run("DefaultEmpty", func(t *testing.T) {
				app := NewAppContext(WithBuildMeta(Meta{Name: "TEST"}))
				defer os.RemoveAll(filepath.Join(app.MainRoot(), "config"))
				v := &struct{ IntArray [3]int }{}
				app.Conf(context.Background(), v)
				Expect(t, v.IntArray, Equal([3]int{}))
			})
			t.Run("HasDefaultValue", func(t *testing.T) {
				app := NewAppContext(WithBuildMeta(Meta{Name: "TEST"}))
				defer os.RemoveAll(filepath.Join(app.MainRoot(), "config"))
				v := &struct{ IntArray [3]int }{IntArray: [3]int{1, 2, 3}}
				app.Conf(context.Background(), v)
				Expect(t, v.IntArray, Equal([3]int{1, 2, 3}))
			})
		})
	})
	t.Run("ConfSimpleMap", func(t *testing.T) {
		t.Run("HasEnvVars", func(t *testing.T) {
			t.Setenv("TEST__SimpleMap1_1", "1")
			t.Setenv("TEST__SimpleMap1_2", "2")
			t.Setenv("TEST__SimpleMap2_a", "10")
			t.Setenv("TEST__SimpleMap2_b", "20")
			t.Run("DefaultEmpty", func(t *testing.T) {
				app := NewAppContext(WithBuildMeta(Meta{Name: "TEST"}))
				defer os.RemoveAll(filepath.Join(app.MainRoot(), "config"))
				v := &struct {
					SimpleMap1 map[int]string
					SimpleMap2 map[string]int
				}{}
				app.Conf(context.Background(), v)
				Expect(t, v.SimpleMap1, Equal(map[int]string{1: "1", 2: "2"}))
				Expect(t, v.SimpleMap2, Equal(map[string]int{"a": 10, "b": 20}))
			})
			t.Run("HasDefaultValue", func(t *testing.T) {
				app := NewAppContext(WithBuildMeta(Meta{Name: "TEST"}))
				defer os.RemoveAll(filepath.Join(app.MainRoot(), "config"))
				v := &struct {
					SimpleMap1 map[int]string
					SimpleMap2 map[string]int
				}{
					SimpleMap1: map[int]string{1: "10", 2: "20", 3: "30"},
					SimpleMap2: map[string]int{"a": 1, "b": 2, "c": 3},
				}
				app.Conf(context.Background(), v)
				Expect(t, v.SimpleMap1, Equal(map[int]string{1: "1", 2: "2", 3: "30"}))
				Expect(t, v.SimpleMap2, Equal(map[string]int{"a": 10, "b": 20, "c": 3}))
			})
		})
		t.Run("NoEnvVars", func(t *testing.T) {
			t.Run("DefaultEmpty", func(t *testing.T) {
				app := NewAppContext(WithBuildMeta(Meta{Name: "TEST"}))
				defer os.RemoveAll(filepath.Join(app.MainRoot(), "config"))
				v := &struct {
					SimpleMap1 map[int]string
					SimpleMap2 map[string]int
				}{}
				Expect(t, v.SimpleMap1, BeNil[map[int]string]())
				Expect(t, v.SimpleMap2, BeNil[map[string]int]())
				app.Conf(context.Background(), v)
				Expect(t, v.SimpleMap1, HaveLen[map[int]string](0))
				Expect(t, v.SimpleMap2, HaveLen[map[string]int](0))
			})
			t.Run("HasDefaultValue", func(t *testing.T) {
				app := NewAppContext(WithBuildMeta(Meta{Name: "TEST"}))
				defer os.RemoveAll(filepath.Join(app.MainRoot(), "config"))
				v := &struct {
					SimpleMap1 map[int]string
					SimpleMap2 map[string]int
				}{
					SimpleMap1: map[int]string{1: "10", 2: "20", 3: "30"},
					SimpleMap2: map[string]int{"a": 1, "b": 2, "c": 3},
				}
				app.Conf(context.Background(), v)
				Expect(t, v.SimpleMap1, Equal(map[int]string{1: "10", 2: "20", 3: "30"}))
				Expect(t, v.SimpleMap2, Equal(map[string]int{"a": 1, "b": 2, "c": 3}))
			})
		})
	})
	t.Run("ConfComplexMap", func(t *testing.T) {
		t.Run("HasEnvVars", func(t *testing.T) {
			t.Setenv("TEST__CONFIG__Map_1_Number", "100")
			t.Setenv("TEST__CONFIG__Map_1_String", "100")
			t.Setenv("TEST__CONFIG__Map_2_Number", "200")
			t.Setenv("TEST__CONFIG__Map_2_String", "200")
			t.Run("DefaultEmpty", func(t *testing.T) {
				app := NewAppContext(WithBuildMeta(Meta{Name: "TEST"}))
				defer os.RemoveAll(filepath.Join(app.MainRoot(), "config"))
				v := &config{}
				app.Conf(context.Background(), v)
				Expect(t, v.Map["1"].Number.Int64(), Equal(int64(100)))
				Expect(t, v.Map["1"].String, Equal("100"))
				Expect(t, v.Map["2"].Number.Int64(), Equal(int64(200)))
				Expect(t, v.Map["2"].String, Equal("200"))
			})
			t.Run("HasDefaultValue", func(t *testing.T) {
				app := NewAppContext(WithBuildMeta(Meta{Name: "TEST"}))
				defer os.RemoveAll(filepath.Join(app.MainRoot(), "config"))
				v := &config{
					Map: map[string]*MapConfig{
						"1": {Number: big.NewInt(1), String: "1"},
						"2": {Number: big.NewInt(2), String: "2"},
						"3": {Number: big.NewInt(300), String: "300"},
					},
				}
				app.Conf(context.Background(), v)
				Expect(t, v.Map["1"].Number.Int64(), Equal(int64(100)))
				Expect(t, v.Map["1"].String, Equal("100"))
				Expect(t, v.Map["2"].Number.Int64(), Equal(int64(200)))
				Expect(t, v.Map["2"].String, Equal("200"))
				Expect(t, v.Map["3"].Number.Int64(), Equal(int64(300)))
				Expect(t, v.Map["3"].String, Equal("300"))
			})
		})
		t.Run("NoEnvVars", func(t *testing.T) {
			t.Run("HasDefaultValue", func(t *testing.T) {
				app := NewAppContext(WithBuildMeta(Meta{Name: "TEST"}))
				defer os.RemoveAll(filepath.Join(app.MainRoot(), "config"))
				v := &config{
					Map: map[string]*MapConfig{
						"k1": {
							Number: big.NewInt(10),
							String: "10",
						},
						"k2": {
							Number: big.NewInt(20),
							String: "20",
						},
					},
				}
				app.Conf(context.Background(), v)
				Expect(t, v.Map["k1"].Number.Int64(), Equal(int64(10)))
				Expect(t, v.Map["k1"].String, Equal("10"))
				Expect(t, v.Map["k2"].Number.Int64(), Equal(int64(20)))
				Expect(t, v.Map["k2"].String, Equal("20"))
			})
			t.Run("InvalidKeyType", func(t *testing.T) {
				type invalid struct {
					Map map[*string]MapConfig
				}
				v := &invalid{}
				app := NewAppContext(WithBuildMeta(Meta{Name: "TEST"}))
				defer os.RemoveAll(filepath.Join(app.MainRoot(), "config"))
				ExpectPanic[error](
					t, func() { app.Conf(context.Background(), v) },
					IsCodeError(envx.CODE__DEC_INVALID_MAP_KEY_TYPE),
				)
			})
			t.Run("InvalidKeyValue", func(t *testing.T) {
				v := &config{
					Map: map[string]*MapConfig{"_key__": nil},
				}
				app := NewAppContext(WithBuildMeta(Meta{Name: "TEST"}))
				defer os.RemoveAll(filepath.Join(app.MainRoot(), "config"))
				ExpectPanic[error](
					t, func() { app.Conf(context.Background(), v) },
					IsCodeError(envx.CODE__ENC_INVALID_MAP_KEY_VALUE),
				)
			})
		})
	})
}
