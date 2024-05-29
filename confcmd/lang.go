package confcmd

type LangType string

const (
	LangEN LangType = "en"
	LangZH LangType = "zh"
)

const DefaultLang = LangEN

var MultiLangHelpKeys = []string{FlagHelp, LangEN.HelpKey(), LangZH.HelpKey()}

func (lang LangType) HelpKey() string {
	return FlagHelp + "." + string(lang)
}
