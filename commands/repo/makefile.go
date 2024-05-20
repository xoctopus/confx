package repo

import "github.com/spf13/cobra"

type GoRepositoryMakefileOption struct {
	EnableCgo       bool     `cmd:""`
	CLibDirectories []string `cmd:""`
	CLibNames       []string `cmd:""`
	OutputDIR       string   `cmd:""`
}

func (v *GoRepositoryMakefileOption) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "makefile",
		Short: "generate go repository makefile",
	}
	return cmd
}
