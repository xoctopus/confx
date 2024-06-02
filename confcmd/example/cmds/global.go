package cmds

// Global defines the command's global flags
type Global struct {
	Debug    bool   `cmd:",p,nop=1"    help:"debug mode"`
	LogLevel string `cmd:",p"          help:"set log level [trace debug info warn error]"`
	Security bool   `cmd:",p,nop=true" help:"enable https serve and request"`
	CertFile string `cmd:",p"          help:"https cert file path"`
	KeyFile  string `cmd:",p"          help:"https key file path"`
}

var DefaultGlobal = &Global{
	Debug:    false,
	LogLevel: "debug",
	Security: false,
}

func (g *Global) SecurityEnabled() bool {
	return g != nil && g.Security && g.CertFile != "" && g.KeyFile != ""
}
