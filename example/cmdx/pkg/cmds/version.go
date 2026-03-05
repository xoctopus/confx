package cmds

import (
	"github.com/spf13/cobra"

	"cgtech.gitlab.com/saitox/confx/pkg/cmdx"
)

// Version output command version
type Version struct{}

func (*Version) Exec(cmd *cobra.Command, args ...string) error {
	cmd.Println("v0.0.1")
	return nil
}

var CmdVersion = cmdx.NewCommand("version", &Version{}).Cmd()
