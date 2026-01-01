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
	root.AddCommand(cmds.ServerCmd.Cmd())
	root.AddCommand(cmds.VersionCmd.Cmd())
}

func main() {
	if err := root.Execute(); err != nil {
		root.PrintErr(err)
	}
}
