package main

import (
	"github.com/spf13/cobra"
	"github.com/xoctopus/x/misc/must"

	"github.com/xoctopus/confx/confcmd"
	"github.com/xoctopus/confx/confcmd/example/cmds"
)

var (
	root *cobra.Command
)

func init() {
	root = &cobra.Command{}

	globals := confcmd.ParseFlags(cmds.DefaultGlobal)
	for _, f := range globals {
		must.NoErrorWrap(
			f.Register(root, confcmd.LangEN, ""),
			"failed to register global flag: %s", f.Name(),
		)
	}

	root.AddCommand(cmds.ServerCmd)
}

func main() {
	if err := root.Execute(); err != nil {
		root.PrintErr(err)
	}
}
