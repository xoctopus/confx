package main

import (
	"github.com/spf13/cobra"

	"github.com/xoctopus/confx/example/cmdx/pkg/cmds"
)

var (
	root *cobra.Command
)

func init() {
	root = &cobra.Command{}
	root.AddCommand(cmds.CmdVersion)
	root.AddCommand(cmds.CmdServer)
}

func main() {
	if err := root.Execute(); err != nil {
		root.PrintErr(err)
	}
}
