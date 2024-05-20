package cmd

import "github.com/spf13/cobra"

type GoCmdMakefileGenerator struct {
	EnableCompileOptimize bool     `cmd:""`
	UseStaticBuildMode    bool     `cmd:""`
	WithImageEntry        bool     `cmd:""`
	WithBuildMetaFlags    bool     `cmd:""`
	BuildMetaPath         string   `cmd:""`
	EnableCgo             bool     `cmd:""`
	CLibDirectories       []string `cmd:""`
	CLibNames             []string `cmd:""`
	TargetName            string   `cmd:""`
	OutputDIR             string   `cmd:""`
}

func (v *GoCmdMakefileGenerator) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "makefile",
		Short: "generate go cmd makefile",
	}
	return cmd
}
