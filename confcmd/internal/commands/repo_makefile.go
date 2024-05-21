package commands

type GoRepositoryMakefileOption struct {
	EnableCgo       bool     `cmd:""`
	CLibDirectories []string `cmd:""`
	CLibNames       []string `cmd:""`
	OutputDIR       string   `cmd:""`
}
