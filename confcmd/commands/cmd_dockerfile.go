package commands

type GoCmdDockerfileGenerator struct {
	Platform            string `cmd:""`
	BuilderImageVersion string `cmd:""`
	RuntimeImageName    string `cmd:""`
	RuntimeImageVersion string `cmd:""`
	Expose              uint16 `cmd:""`
	TargetName          string `cmd:""`
	OutputDIR           string `cmd:""`
	// todo if need certification support for tls/ssl https
}
