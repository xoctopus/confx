package repo

import "github.com/spf13/cobra"

type GoGithubCIWorkflowOption struct {
	Branches            []string
	RunsOn              string
	WithTesting         bool
	WithTargetsBuilding bool
	WithCoverageReport  bool
}

func (v *GoGithubCIWorkflowOption) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dockerfile",
		Short: "generate go github ci workflow",
	}
	return cmd
}
