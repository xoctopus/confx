package main

import (
	"github.com/spf13/cobra"

	"github.com/xoctopus/confx/confcmd/example/cmds"
)

var (
	root *cobra.Command
)

func init() {
	root = &cobra.Command{}
	root.AddCommand(cmds.ServerCmd)
	root.AddCommand(cmds.VersionCmd)
}

func main() {
	if err := root.Execute(); err != nil {
		root.PrintErr(err)
	}
}
