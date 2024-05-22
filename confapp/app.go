package confapp

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"

	"github.com/sincospro/x/misc/must"
	"github.com/sincospro/x/reflectx"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/sincospro/conf/confapp/initializer"
	"github.com/sincospro/conf/confcmd"
	"github.com/sincospro/conf/confcmd/commands"
	"github.com/sincospro/conf/envconf"
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
		app.option.Meta.Overwrite(meta)
	}
}

func WithDefaultConfigGenerator() Option {
	return func(app *AppCtx) {
		app.option.GenDefaults = true
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
				cmd.Println(app.Version())

				sort.Slice(app.vars, func(i, j int) bool {
					return app.vars[i].Name < app.vars[j].Name
				})
				for _, g := range app.vars {
					cmd.Println(string(g.MaskBytes()))
				}
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
	root   string           // root main.go path
	dfts   []*envconf.Group // dfts app default config var groups
	vars   []*envconf.Group // vars app config var groups
	option AppOption        // option application options
}

func (app *AppCtx) Name() string {
	return app.option.Meta.Name
}

func (app *AppCtx) Version() string {
	return app.option.Meta.String()
}

func (app *AppCtx) Conf(configurations ...any) {
	app.injectLocalConfig()

	app.dfts = make([]*envconf.Group, 0, len(configurations))
	app.vars = make([]*envconf.Group, 0, len(configurations))
	vars := make([]reflect.Value, 0, len(configurations))

	for _, c := range configurations {
		rv := reflect.ValueOf(c)
		group := app.group(rv)

		app.dfts = append(app.dfts, app.marshalDefaults(group, rv))
		app.vars = append(app.vars, app.scanEnvironment(group, rv))
		vars = append(vars, rv)
	}

	app.initial(vars)
	app.attachSubCommands()
}

// injectLocalConfig try parse vars in local.yaml, and inject vars to environment
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

// marshalDefaults encode default vars
func (app *AppCtx) marshalDefaults(group string, v any) *envconf.Group {
	dft := envconf.NewGroup(group)
	must.NoError(envconf.NewDecoder(dft).Decode(v))
	must.NoError(envconf.NewEncoder(dft).Encode(v))
	return dft
}

// scanEnvironment scan vars from environment
func (app *AppCtx) scanEnvironment(group string, v any) *envconf.Group {
	vars := envconf.ParseGroupFromEnv(group)
	must.NoError(envconf.NewDecoder(vars).Decode(v))
	return vars
}

func initialize(v reflect.Value) {
	if initializer.CanBeInitialized(v) {
		must.NoError(initializer.Init(v))
		return
	}
	v = reflectx.Indirect(v)
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			initialize(v.Field(i))
		}
	}
}

func (app *AppCtx) initial(vars []reflect.Value) {
	for i := range vars {
		initialize(vars[i])
	}
}

func (app *AppCtx) group(rv reflect.Value) string {
	group := rv.Type().Name()
	rv = reflectx.Indirect(rv)
	group = rv.Type().Name()
	if group == "" {
		return strings.ToUpper(strings.Replace(app.Name(), "-", "_", -1))
	}
	return strings.ToUpper(strings.Replace(app.Name()+"__"+group, "-", "_", -1))
}

func (app *AppCtx) attachSubCommands() {
	if !app.option.NeedAttach() {
		return
	}
	gen := &cobra.Command{
		Use:   "gen",
		Short: "generator templates files for makefile, dockerfile and default config",
	}
	app.Command.AddCommand(gen)

	if app.option.GenDefaults {
		option := &commands.GoCmdGenDefaultConfigOptions{
			Defaults: &app.dfts,
			Output:   filepath.Join(app.root, "config"),
		}
		cmd, err := confcmd.NewCommand(confcmd.EN, option)
		must.NoError(err)
		gen.AddCommand(cmd)
	}

	if app.option.GenDockerfile {
	}

	if app.option.GenMakefile {
	}
}
