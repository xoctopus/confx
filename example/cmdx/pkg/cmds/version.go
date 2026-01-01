package cmds

import (
	"github.com/spf13/cobra"

	"github.com/xoctopus/confx/pkg/cmdx"
)

// Version output command version
type Version struct{}

func (*Version) Exec(cmd *cobra.Command, args ...string) error {
	cmd.Println("v0.0.1")
	return nil
}

var VersionCmd = cmdx.NewCommand("version", &Version{}).
	WithShort("output command version")
