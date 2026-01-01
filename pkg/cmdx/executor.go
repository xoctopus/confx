package cmdx

import (
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/stringsx"

	"github.com/xoctopus/confx/pkg/envx"
)

// Command used to generate a cobra.Command
//
// it defines the basic info for constructing cobra.Command, such as usage
// message, short description and execution method
type Command interface {
	// Use command usage message, as `cobra.Command.Use`
	Use() string
	// WithShort adds short description of this executor
	WithShort(string) Command
	// WithLong adds long descriptions
	WithLong(string) Command
	// WithExample adds example description
	WithExample(string) Command
	// Cmd returns generated *cobra.Command
	Cmd() *cobra.Command
}

// Executor defines executor handling logic
type Executor interface {
	Exec(cmd *cobra.Command, args ...string) error
}

func NewCommand(use string, executor Executor) Command {
	must.BeTrueF(
		stringsx.ValidIdentifier(use),
		"invalid command use, it should be a valid identifier",
	)

	e := &command{
		use: use,
		env: strings.ToUpper(use),
	}

	flags := parseFlags(
		reflect.ValueOf(executor),
		envx.NewPathWalker(),
		map[string]map[string]*Flag{
			"env":   {},
			"cmd":   {},
			"short": {},
		},
	)

	e.cmd = &cobra.Command{
		Use: use,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executor.Exec(cmd, args...)
		},
	}

	for _, f := range flags {
		f.Register(e.cmd, use)
	}

	return e
}

type command struct {
	use string
	env string
	cmd *cobra.Command
}

func (c *command) Use() string {
	return c.use
}

func (c *command) WithShort(short string) Command {
	c.cmd.Short = short
	return c
}

func (c *command) WithLong(v string) Command {
	c.cmd.Long = v
	return c
}

func (c *command) WithExample(v string) Command {
	c.cmd.Example = v
	return c
}

func (c *command) Cmd() *cobra.Command {
	return c.cmd
}
