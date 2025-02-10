package cmds

import (
	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "print command version",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("v0.0.1")
	},
}
