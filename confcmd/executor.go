package confcmd

import (
	"sort"

	"github.com/modood/table"
	"github.com/spf13/cobra"
	"github.com/xoctopus/x/misc/must"
)

type Executor interface {
	// Use command usage message, as `cobra.Command.Use`
	Use() string
	// Short a short description of this executor
	Short() string
	// Exec impls executor handle logic
	Exec(cmd *cobra.Command) error
	// Flag returns executor's flag by flag name
	Flag(name string) *Flag
	// Flags returns executor's flag map
	Flags() map[string]*Flag
	// AddFlag add flag to executor
	AddFlag(*Flag)
	// OutputFlagsHelp output flags help
	OutputFlagsHelp(lang LangType, prefix string)
	// HelpLang returns support help language type
	HelpLang() LangType
	// SetHelpLang set help language
	SetHelpLang(lang LangType)
}

type WithLong interface {
	// Long executor's long help message
	Long() string
}

type WithExample interface {
	// Example examples of how to use
	Example() string
}

func NewFlagSet() *FlagSet {
	return &FlagSet{
		flags: map[string]*Flag{},
		envs:  map[string]*Flag{},
	}
}

// FlagSet supports management of flags and envs
type FlagSet struct {
	flags map[string]*Flag
	envs  map[string]*Flag
}

func (fs *FlagSet) Flag(name string) *Flag {
	return fs.flags[name]
}

func (fs *FlagSet) Flags() map[string]*Flag {
	return fs.flags
}

func (fs *FlagSet) AddFlag(f *Flag) {
	name := f.Name()
	_, ok := fs.flags[name]
	must.BeTrueWrap(!ok, "flag name conflict: %s", name)
	fs.flags[name] = f

	if key := f.EnvKey(""); key != "" {
		_, ok = fs.envs[key]
		must.BeTrueWrap(!ok, "env key conflict: %s", key)
		fs.envs[key] = f
	}
}

type Output struct {
	FlagName string `table:"flag name"`
	Required string `table:"required"`
	Default  any    `table:"default value"`
	EnvKey   string `table:"environment key"`
	Help     string `table:"help info"`
}

func (fs *FlagSet) OutputFlagsHelp(lang LangType, prefix string) {
	flags := make([]*Flag, 0, len(fs.flags))
	for _, f := range fs.flags {
		flags = append(flags, f)
	}
	sort.Slice(flags, func(i, j int) bool {
		return flags[i].name < flags[j].name
	})
	output := make([]Output, 0, len(flags))
	for _, f := range flags {
		o := Output{
			FlagName: f.Name(),
			Default:  f.DefaultValue(),
			EnvKey:   f.EnvKey(prefix),
			Help:     f.Help(lang),
		}
		if f.IsRequired() {
			o.Required = "yes"
		}
		output = append(output, o)
	}
	table.Output(output)
}

func NewMultiLangHelper(lang LangType) *MultiLangHelper {
	return &MultiLangHelper{lang: lang}
}

func NewDefaultMultiLangHelper() *MultiLangHelper {
	return NewMultiLangHelper(DefaultLang)
}

// MultiLangHelper support executor's flag has multi-language help
type MultiLangHelper struct {
	lang LangType
}

func (l *MultiLangHelper) HelpLang() LangType {
	return l.lang
}

func (l *MultiLangHelper) SetHelpLang(lang LangType) {
	l.lang = lang
}

type CanInjectFromEnv interface {
	Prefix() string
	SetPrefix(string)
}

func NewEnvInjector(prefix string) *EnvInjector {
	return &EnvInjector{prefix: prefix}
}

// EnvInjector support executor's flag value to inject from env vars
type EnvInjector struct {
	prefix string
}

func (v *EnvInjector) Prefix() string {
	return v.prefix
}

func (v *EnvInjector) SetPrefix(prefix string) {
	v.prefix = prefix
}
