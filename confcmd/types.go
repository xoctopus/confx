package confcmd

import "github.com/spf13/cobra"

type Executor interface {
	Use() string
	Short() string
	Exec(cmd *cobra.Command) error
}

type WithLong interface {
	Long() string
}

type WithExample interface {
	Example()
}

type Lang string

const (
	EN = "en"
	ZH = "zh"
)

const (
	HelpKey   = "help"
	HelpEnKey = "help.en"
	HelpZhKey = "help.zh"
)

const (
	flagKey      = "name"
	requiredFlag = "required"
)

var (
	helpKeys = []string{HelpKey, HelpEnKey, HelpZhKey}
)
