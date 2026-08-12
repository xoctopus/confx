package appx

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/reflectx"
	"gopkg.in/yaml.v3"

	"github.com/xoctopus/confx/pkg/envx"
	"github.com/xoctopus/confx/pkg/types"
)

// Option configures an [AppCtx] at construction time.
type Option func(*AppCtx)

// WithMainRoot sets the application root used to resolve config/local.yml,
// config/default.yml, and config/.env. When empty, [NewAppContext] keeps the
// process working directory.
func WithMainRoot(root string) Option {
	return func(app *AppCtx) {
		app.root = root
	}
}

// WithBuildMeta sets the application build [Meta], replacing [DefaultMeta].
func WithBuildMeta(meta Meta) Option {
	return func(app *AppCtx) {
		app.option.Meta = new(meta)
	}
}

// WithPreRunner appends ordered callbacks run before Serves on `run`.
//
// Use for sequential startup initialization such as loading global config or
// injecting a global context, before long-lived services start.
func WithPreRunner(runners ...func()) Option {
	return func(app *AppCtx) {
		app.option.PreRunners = append(app.option.PreRunners, runners...)
	}
}

// WithServes appends parallel callbacks run after PreRunners on `run`.
//
// Use for long-lived services such as HTTP servers or scheduled jobs.
func WithServes(serves ...func()) Option {
	return func(app *AppCtx) {
		app.option.Serves = append(app.option.Serves, serves...)
	}
}

// WithCloseFns registers extra close callbacks invoked by [AppCtx.Close].
//
// Closable components loaded via [AppCtx.Conf] do not need registration; they
// are closed automatically when [AppCtx.Close] runs.
func WithCloseFns(closes ...func() error) Option {
	return func(app *AppCtx) {
		app.option.CloseFns = append(app.option.CloseFns, closes...)
	}
}

// NewAppContext builds an [AppCtx] with main as the `run` entry, a hidden root
// cobra command, and a `version` subcommand. Apply options before [AppCtx.Conf].
func NewAppContext(main func() error, options ...Option) *AppCtx {
	app := &AppCtx{
		cmd:    &cobra.Command{},
		root:   must.NoErrorV(os.Getwd()),
		option: AppOption{Meta: new(DefaultMeta)},
	}

	app.AddCommand(
		"run",
		"run app's main entry",
		func( /*cmd *cobra.Command,*/ args []string) error {
			fmt.Printf("%s\n\n", color.HiCyanString(app.Version()))
			app.log()
			app.option.PreRun()
			return main()
		},
	)

	for _, opt := range options {
		opt(app)
	}

	app.cmd.Use = app.Name()
	app.cmd.Hidden = true
	app.cmd.CompletionOptions.DisableDefaultCmd = true
	app.cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "display app version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(app.Version())
		},
	})

	return app
}

// AppCtx is the application handle: cobra CLI, build meta, and env config groups.
type AppCtx struct {
	cmd *cobra.Command

	root       string
	dfts       []*envx.Group
	vars       []*envx.Group
	components []reflect.Value

	option AppOption
}

// Name returns the application name from build [Meta].
func (app *AppCtx) Name() string {
	return app.option.Meta.Name
}

// Version returns the formatted build identity string from [Meta.String].
func (app *AppCtx) Version() string {
	return app.option.Meta.String()
}

// MainRoot returns the path set by [WithMainRoot], or the working directory
// used when constructing the app.
func (app *AppCtx) MainRoot() string {
	return app.root
}

// Conf loads one or more config pointers from the environment, writes default
// config files under MainRoot/config, and initializes fields that support
// types.InitByContext. Meta is injected into ctx via [WithAppMeta] before init.
//
// Each named config type becomes an envx group (e.g. APP__OTEL). Anonymous
// structs are allowed only when a single configuration is passed.
func (app *AppCtx) Conf(ctx context.Context, configurations ...any) context.Context {
	app.injectLocalConfig()

	app.dfts = make([]*envx.Group, 0, len(configurations))
	app.vars = make([]*envx.Group, 0, len(configurations))
	app.components = make([]reflect.Value, 0, len(configurations))
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
		app.components = append(app.components, rv)
	}

	app.mustWriteDefault()

	return app.initial(WithAppMeta(ctx, *app.option.Meta))
}

// Close shuts down the application.
//
// It first closes components registered through [AppCtx.Conf] via
// types.CloseByContext (no [WithCloseFns] needed for those), then runs any
// callbacks registered with [WithCloseFns]. Errors are joined and returned.
func (app *AppCtx) Close(ctx context.Context) error {
	errs := make([]error, 0, len(app.components))
	for i := range app.components {
		if err := types.CloseByContext(ctx, app.components[i]); err != nil {
			errs = append(errs, err)
		}
	}
	for i := range app.option.CloseFns {
		if f := app.option.CloseFns[i]; f != nil {
			if err := f(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Execute runs the root cobra command (typically from main).
//
// Dispatches to registered subcommands such as `run` / `version`, or any
// command added via [AppCtx.AddCommand].
func (app *AppCtx) Execute() error {
	return app.cmd.Execute()
}

// AddCommand registers a cobra subcommand under the app root.
//
// name is the command Use string, short is the Short help text, and runE is
// invoked with the remaining CLI args when the command is selected.
// [NewAppContext] uses this to register `run`.
func (app *AppCtx) AddCommand(name, short string, runE func(args []string) error) {
	app.cmd.AddCommand(
		&cobra.Command{
			Use:   name,
			Short: short,
			RunE: func(_ *cobra.Command, args []string) error {
				return runE(args)
			},
		},
	)
}

func (app *AppCtx) injectLocalConfig() {
	local, err := os.ReadFile(filepath.Join(app.root, "./config/local.yml"))
	if err == nil {
		kv := make(map[string]string)
		if err = yaml.Unmarshal(local, &kv); err == nil {
			for k, v := range kv {
				if _, ok := os.LookupEnv(k); !ok {
					_ = os.Setenv(k, v)
				}
			}
		}
	}
}

func (app *AppCtx) marshalDefaults(group string, v any) *envx.Group {
	dft := envx.NewGroup(group)
	must.NoErrorF(envx.NewDecoder(dft).Decode(v), "failed to decode default")
	must.NoErrorF(envx.NewEncoder(dft).Encode(v), "failed to encode default")
	return dft
}

func (app *AppCtx) scanEnvironment(group string, v any) *envx.Group {
	vars := envx.ParseGroupFromEnv(group)
	must.NoErrorF(envx.NewDecoder(vars).Decode(v), "failed to decode env")
	must.NoErrorF(envx.NewEncoder(vars).Encode(v), "failed to encode env")
	return vars
}

func initialize(ctx context.Context, v reflect.Value, g *envx.Group, field string) context.Context {
	err := types.InitByContext(ctx, v)
	if errors.Is(err, types.ErrSkipInitializing) {
		return ctx
	}
	must.NoErrorF(err, "failed to init [group:%s] [field:%s]", g.Name(), field)

	ctx = types.Inject(ctx, v)

	v = reflectx.Indirect(v)
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() {
				ctx = initialize(ctx, v.Field(i), g, v.Type().Field(i).Name)
			}
		}
	}
	return ctx
}

func (app *AppCtx) initial(ctx context.Context) context.Context {
	for i := range app.components {
		ctx = initialize(ctx, app.components[i], app.vars[i], "")
	}
	return ctx
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

	m = make(map[string]string)
	for _, g := range app.dfts {
		for _, v := range g.Values() {
			if !v.Optional() {
				m[g.Key(v.Key())] = v.Value()
			}
		}
	}

	content = envx.DotEnv(m)
	filename = filepath.Join(dir, ".env")
	must.NoErrorF(
		os.WriteFile(filename, content, os.ModePerm),
		"failed to write default config file",
	)
}
