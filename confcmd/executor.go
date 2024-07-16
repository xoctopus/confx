package confcmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// Executor used to generate a cobra.Command
//
// it defines the basic info for constructing cobra.Command, such as usage
// message, short description and execution method
type Executor interface {
	// Use command usage message, as `cobra.Command.Use`
	Use() string
	// Short a short description of this executor
	Short() string
	// Exec impls executor handle logic
	Exec(cmd *cobra.Command, args ...string) error
}

// WithLong defines executor's lang help message
type WithLong interface {
	Long() string
}

// WithExample defines executor's examples of how to use
type WithExample interface {
	Example() string
}

// WithEnv if an executor impls `WithEnv`, it allows executor's flag
// to be parsed from env vars
type WithEnv interface {
	Prefix() string
	Var(key string) string
}

func NewEnv(prefix string) *Env {
	return &Env{prefix: prefix}
}

type Env struct {
	prefix string
}

func (v *Env) Prefix() string { return v.prefix }

func (v *Env) Var(key string) string {
	if key == "" {
		return ""
	}
	envKey := ""
	if v.prefix != "" {
		envKey = strings.Replace(strings.ToUpper(v.prefix+"__"+key), "-", "_", -1)
	} else {
		envKey = strings.Replace(strings.ToUpper(key), "-", "_", -1)
	}
	return os.Getenv(envKey)
}

// LocalizeHelper help translate flag help message by i18n message id
type LocalizeHelper func(i18nId string) string

// NewCommand generate a `*cobra.Command` by Executor
func NewCommand(v Executor, localize ...LocalizeHelper) *cobra.Command {
	flags := ParseFlags(v)

	cmd := &cobra.Command{
		Use:   v.Use(),
		Short: v.Short(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Exec(cmd, args...)
		},
	}

	if exec, ok := v.(WithLong); ok {
		cmd.Long = exec.Long()
	}
	if exec, ok := v.(WithExample); ok {
		cmd.Example = exec.Example()
	}

	env, _ := v.(WithEnv)
	trans := LocalizeHelper(nil)
	if len(localize) > 0 && localize[0] != nil {
		trans = localize[0]
	}

	for _, f := range flags {
		envVar := ""
		if env != nil {
			envVar = env.Var(f.Env())
		}
		f.Register(cmd, envVar, trans)
	}

	return cmd
}
