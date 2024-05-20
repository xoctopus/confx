package cmd

import "github.com/spf13/cobra"

type GoCmdDockerfileGenerator struct {
	Platform            string `cmd:""`
	BuilderImageVersion string `cmd:""`
	RuntimeImageName    string `cmd:""`
	RuntimeImageVersion string `cmd:""`
	Expose              uint16 `cmd:""`
	TargetName          string `cmd:""`
	OutputDIR           string `cmd:""`
}

func (v *GoCmdDockerfileGenerator) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dockerfile",
		Short: "generate go cmd dockerfile",
	}
	return cmd
}
