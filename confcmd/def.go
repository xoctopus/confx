package confcmd

const (
	// TAG_CMD presents command flag name
	TAG_CMD = "cmd"
	// TAG_ENV presents command value from env
	TAG_ENV  = "env"
	TAG_I18N = "i18n" // tagI18n presents command flag multi-language message id
	TAG_HELP = "help" // tagHelp presents command flag help usage message
)

const (
	// OPTION_REQUIRE marks flag as required
	OPTION_REQUIRE = "require"
	// OPTION_PERSIST marks flag persistent
	OPTION_PERSIST = "persist"
	// OPTION_DEFAULT marks flag default value, if flag has no option
	OPTION_DEFAULT = "default"
	// OPTION_SHORT marks flag shorthand
	OPTION_SHORT = "short"
)
