package appx

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/xoctopus/x/initializer"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/reflectx"
	"gopkg.in/yaml.v3"

	"github.com/xoctopus/confx/pkg/envx"
)

type Option func(*AppCtx)

func WithMainRoot(root string) Option {
	_, filename, _, _ := runtime.Caller(1)
	return func(app *AppCtx) {
		app.root = filepath.Join(filepath.Dir(filename), root)
	}
}

func WithBuildMeta(meta Meta) Option {
	return func(app *AppCtx) {
		app.option.Meta = DefaultMeta
		app.option.Meta.Overwrite(meta)
	}
}

func WithMakefileGenerator() Option {
	return func(app *AppCtx) {
		app.option.GenMakefile = true
	}
}

func WithDockerfileGenerator() Option {
	return func(app *AppCtx) {
		app.option.GenDockerfile = true
	}
}

func WithMainExecutor(main func() error) Option {
	return func(app *AppCtx) {
		app.Command.AddCommand(&cobra.Command{
			Use:   "run",
			Short: "run app's main entry",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Printf("%s\n\n", color.HiCyanString(app.Version()))
				app.log()
				app.option.PreRun()
				return main()
			},
		})
	}
}

func WithPreRunner(runners ...func()) Option {
	return func(app *AppCtx) {
		app.option.PreRunners = append(app.option.PreRunners, runners...)
	}
}

func WithBatchRunner(runners ...func()) Option {
	return func(app *AppCtx) {
		app.option.BatchRunners = append(app.option.BatchRunners, runners...)
	}
}

func NewAppContext(options ...Option) *AppCtx {
	app := &AppCtx{
		Command: &cobra.Command{},
		option:  AppOption{Meta: DefaultMeta},
	}

	for _, opt := range options {
		opt(app)
	}

	app.Command.Use = app.Name()
	app.Command.Hidden = true
	app.Command.CompletionOptions.DisableDefaultCmd = true
	app.Command.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "display app version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(app.Version())
		},
	})

	return app
}

type AppCtx struct {
	*cobra.Command
	root   string        // root main.go path
	dfts   []*envx.Group // dfts app default config var groups
	vars   []*envx.Group // vars app config var groups
	option AppOption     // option application options
}

func (app *AppCtx) Name() string {
	return app.option.Meta.Name
}

func (app *AppCtx) Version() string {
	return app.option.Meta.String()
}

func (app *AppCtx) MainRoot() string {
	return app.root
}

func (app *AppCtx) Conf(configurations ...any) {
	app.injectLocalConfig()

	app.dfts = make([]*envx.Group, 0, len(configurations))
	app.vars = make([]*envx.Group, 0, len(configurations))
	vars := make([]reflect.Value, 0, len(configurations))
	names := map[string]struct{}{}

	for _, c := range configurations {
		rv := reflect.ValueOf(c)
		name := reflectx.Indirect(rv).Type().Name()

		_, ok := names[name]
		must.BeTrueF(!ok, "config name conflicted")

		if len(configurations) > 1 {
			must.BeTrueF(name != "", "anonymous config when more than one")
		}

		group := app.group(name)

		app.dfts = append(app.dfts, app.marshalDefaults(group, rv))
		app.vars = append(app.vars, app.scanEnvironment(group, rv))
		vars = append(vars, rv)
	}

	app.mustWriteDefault()
	app.initial(vars)
	if app.option.NeedAttach() {
		app.attachSubCommands()
	}
}

// injectLocalConfig try parse vars in local.yaml, and inject vars to environment
func (app *AppCtx) injectLocalConfig() {
	local, err := os.ReadFile(filepath.Join(app.root, "./config/local.yml"))
	if err == nil {
		kv := make(map[string]string)
		if err = yaml.Unmarshal(local, &kv); err == nil {
			for k, v := range kv {
				_ = os.Setenv(k, v)
			}
		}
	}
}

// marshalDefaults encode default vars
func (app *AppCtx) marshalDefaults(group string, v any) *envx.Group {
	dft := envx.NewGroup(group)
	must.NoErrorF(envx.NewDecoder(dft).Decode(v), "failed to decode default")
	must.NoErrorF(envx.NewEncoder(dft).Encode(v), "failed to encode default")
	return dft
}

// scanEnvironment scan vars from environment
func (app *AppCtx) scanEnvironment(group string, v any) *envx.Group {
	vars := envx.ParseGroupFromEnv(group)
	must.NoErrorF(envx.NewDecoder(vars).Decode(v), "failed to decode env")
	must.NoErrorF(envx.NewEncoder(vars).Encode(v), "failed to encode env")
	return vars
}

func initialize(v reflect.Value, g *envx.Group, field string) {
	if initializer.CanBeInitialized(v) {
		must.NoErrorF(
			initializer.Init(v),
			"failed to init [group:%s] [field:%s]", g.Name(), field,
		)
		return
	}
	v = reflectx.Indirect(v)
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() {
				initialize(v.Field(i), g, v.Type().Field(i).Name)
			}
		}
	}
}

func (app *AppCtx) initial(vars []reflect.Value) {
	for i := range vars {
		initialize(vars[i], app.vars[i], "")
	}
}

func (app *AppCtx) log() {
	app.option.Meta.Print()

	sort.Slice(app.vars, func(i, j int) bool {
		return app.vars[i].Name() < app.vars[j].Name()
	})

	for i := range app.vars {
		fmt.Print(color.HiBlueString("%s", app.vars[i].MaskBytes()))
	}
	fmt.Println("")
}

func (app *AppCtx) group(name string) string {
	if name == "" {
		return strings.ToUpper(strings.Replace(app.Name(), "-", "_", -1))
	}
	return strings.ToUpper(strings.Replace(app.Name()+"__"+name, "-", "_", -1))
}

func (app *AppCtx) mustWriteDefault() {
	dir := filepath.Join(app.root, "config")

	must.NoErrorF(
		os.MkdirAll(dir, os.ModePerm),
		"failed to create output dir",
	)

	m := make(map[string]string)
	for _, g := range app.dfts {
		for _, v := range g.Values() {
			if !v.Optional() {
				m[g.Key(v.Key())] = v.Value()
			}
		}
	}

	content, err := yaml.Marshal(m)
	must.NoErrorF(err, "failed to marshal default vars")

	filename := filepath.Join(dir, "default.yml")
	must.NoErrorF(
		os.WriteFile(filename, content, os.ModePerm),
		"failed to write default config file",
	)
}

func (app *AppCtx) attachSubCommands() {
	gen := &cobra.Command{
		Use:   "gen",
		Short: "generator templates files for makefile, dockerfile and default config",
	}

	if app.option.GenDockerfile {
	}

	if app.option.GenMakefile {
	}

	app.Command.AddCommand(gen)
}
