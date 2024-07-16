package confcmd

const (
	tagCmd  = "cmd"  // tagCmd presents command flag name
	tagEnv  = "env"  // tagEnv presents command value from env
	tagI18n = "i18n" // tagI18n presents command flag multi-language message id
	tagHelp = "help" // tagHelp presents command flag help usage message
)

const (
	flagRequire = "require" // flagRequire marks flag as required
	flagPersist = "persist" // flagRequire marks flag persistent
	flagDefault = "default" // flagDefault marks flag default value, if flag has no option
	flagInline  = "inline"  // flagInline marks field as inline single flag
	flagShort   = "short"   // flagShort marks flag shorthand
)
