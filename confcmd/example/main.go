package main

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/xoctopus/confx/confcmd"
)

type Executor struct {
	ServerExpose   uint16 `cmd:"" help:"server expose port"`
	PrivateKey     string `cmd:"" help:"server private key"`
	ClientEndpoint string `cmd:"" help:"client endpoint"`
	Key            string `cmd:"key,required" help:"query command flag by key"`

	// Executor enhancements
	*confcmd.FlagSet         `cmd:"-"`
	*confcmd.MultiLangHelper `cmd:"-"`
	*confcmd.EnvInjector     `cmd:"-"`
}

var _ confcmd.Executor = (*Executor)(nil)

func (c *Executor) Use() string {
	return "example"
}

func (c *Executor) Short() string {
	return "a cmd factory example"
}

func (c *Executor) Exec(cmd *cobra.Command) error {
	if strings.ToLower(c.Key) == "all" {
		for name, flag := range c.Flags() {
			cmd.Printf("%s ==> %v\n", name, flag.Value())
		}
		return nil
	}
	f := c.Flag(c.Key)
	if f == nil {
		cmd.Printf("%s is not found\n", c.Key)
		return nil
	}
	cmd.Printf("%s ==> %v\n", c.Key, f.Value())
	return nil
}

var (
	cmd *cobra.Command
)

func init() {
	cmd = confcmd.NewCommand(&Executor{
		ServerExpose:   10000,
		ClientEndpoint: "localhost:10001",

		FlagSet:         confcmd.NewFlagSet(),
		MultiLangHelper: confcmd.NewDefaultMultiLangHelper(),
		EnvInjector:     confcmd.NewEnvInjector("example"),
	})
}

func main() {
	_ = cmd.Execute()
}
