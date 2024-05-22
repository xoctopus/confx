package commands

type GoGithubCIWorkflowOption struct {
	Branches            []string
	RunsOn              string
	WithTesting         bool
	WithTargetsBuilding bool
	WithCoverageReport  bool
}
