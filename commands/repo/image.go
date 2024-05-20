package repo

import "github.com/spf13/cobra"

type GoGithubImageWorkflowOption struct {
}

func (v *GoGithubImageWorkflowOption) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dockerfile",
		Short: "generate go github image build and push workflow",
	}
	return cmd
}
