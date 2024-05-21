package confapp

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/sincospro/x/reflectx"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/sincospro/conf/envconf"
)

type Meta struct {
	Name     string
	Feature  string
	Version  string
	CommitID string
	Date     string
	Runtime  Runtime
}

var DefaultMeta = Meta{
	Name:     "name",
	Feature:  "main",
	Version:  "v0.0.0",
	CommitID: "commit",
	Date:     time.Now().Format("200601021504"),
	Runtime:  GetRuntime(),
}

func (m *Meta) String() string {
	return fmt.Sprintf("%s:%s@%s#%s_%s(%s)", m.Name, m.Feature, m.Version, m.CommitID, m.Date, m.Runtime)
}

type Option func(*AppCtx)

func WithWorkDir(root string) Option {
	_, filename, _, _ := runtime.Caller(1)
	return func(app *AppCtx) {
		app.root = filepath.Join(filepath.Dir(filename), root)
	}
}

func WithBuildMeta(meta Meta) Option {
	return func(app *AppCtx) {
		if meta.Name != "" {
			app.Meta.Name = meta.Name
		}
		if meta.Feature != "" {
			app.Meta.Feature = meta.Feature
		}
		if meta.Version != "" {
			app.Meta.Version = meta.Version
		}
		if meta.CommitID != "" {
			app.Meta.CommitID = meta.CommitID
		}
		if meta.Date != "" {
			app.Meta.Date = meta.Date
		}
	}
}

func WithMakefileGenerator() Option {
	return func(app *AppCtx) {
		if app.Meta == DefaultMeta || app.root == "" {
			panic("should set app meta and root")
		}
		// app.gen.AddCommand((&cmd.GoCmdMakefileGenerator{}).Command())
	}
}

func WithDefaultConfigGenerator() Option {
	return func(app *AppCtx) {
		if app.Meta == DefaultMeta || app.root == "" {
			panic("should set app meta and root")
		}
		// app.gen.AddCommand((&cmd.GoCmdDefaultConfigGenerator{}).Command())
	}
}

func WithDockerfileGenerator() Option {
	return func(app *AppCtx) {
		if app.Meta == DefaultMeta || app.root == "" {
			panic("should set app meta and root")
		}
		// app.gen.AddCommand((&cmd.GoCmdDockerfileGenerator{}).Command())
	}
}

func NewAppContext(options ...Option) *AppCtx {
	app := &AppCtx{
		Command: &cobra.Command{},
		Meta:    DefaultMeta,
	}

	for _, opt := range options {
		opt(app)
	}

	app.gen = &cobra.Command{
		Use:   "gen",
		Short: "generator templates for makefile, dockerfile, default config and github ci workflow",
	}
	app.Command.AddCommand(app.gen)

	return app
}

type AppCtx struct {
	Command *cobra.Command
	Meta    Meta
	Runtime Runtime // env app runtime
	root    string  // root main.go path
	gen     *cobra.Command
}

func (app *AppCtx) Version() string {
	return app.Meta.String()
}

func (app *AppCtx) Conf(configurations ...any) error {
	app.injectLocalConfig()

	defaults := make([]*envconf.Group, 0, len(configurations))
	for _, c := range configurations {
		rv := reflect.ValueOf(c)
		// todo check valid
		group := app.group(rv)
		v, err := app.marshalDefaultVar(group, rv)
		if err != nil {
			return err
		}
		defaults = append(defaults, v)
		if err = app.parseFromEnvirons(group, rv); err != nil {
			return err
		}
	}

	return nil
}

// injectLocalConfig try parse vars in local.yaml to environment
func (app *AppCtx) injectLocalConfig() {
	local, err := os.ReadFile(filepath.Join(app.root, "./config/local.yml"))
	if err != nil {
		return
	}
	kv := make(map[string]string)
	if err = yaml.Unmarshal(local, &kv); err == nil {
		for k, v := range kv {
			_ = os.Setenv(k, v)
		}
	}
}

// marshalDefaultVar encode default vars
func (app *AppCtx) marshalDefaultVar(group string, v any) (*envconf.Group, error) {
	dft := envconf.NewGroup(group)
	if err := envconf.NewDecoder(dft).Decode(v); err != nil {
		return nil, err
	}
	if err := envconf.NewEncoder(dft).Encode(v); err != nil {
		return nil, err
	}
	return dft, nil
}

// parseFromEnvirons parse vars from environment
func (app *AppCtx) parseFromEnvirons(group string, v any) error {
	vars := envconf.ParseGroupFromEnv(group)
	if err := envconf.NewDecoder(vars).Decode(v); err != nil {
		return err
	}
	return nil
}

func (app *AppCtx) group(rv reflect.Value) string {
	group := rv.Type().Name()
	rv = reflectx.Indirect(rv)
	group = rv.Type().Name()
	if group == "" {
		return strings.ToUpper(strings.Replace(app.Meta.Name, "-", "_", -1))
	}
	return strings.ToUpper(strings.Replace(app.Meta.Name+"__"+group, "-", "_", -1))
}

func (app *AppCtx) log(group string, v any) {
	vars := envconf.NewGroup(group)
	if err := envconf.NewEncoder(vars).Encode(v); err != nil {
		panic(err)
	}
	fmt.Printf("%s", string(vars.Bytes()))
}
