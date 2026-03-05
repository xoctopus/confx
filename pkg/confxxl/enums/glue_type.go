package enums

// GlueType defines glue type of xxl-job trigger
// +genx:enum
type GlueType int8

const (
	GLUE_TYPE_UNKNOWN GlueType = iota
	GLUE_TYPE__BEAN
	GLUE_TYPE__GLUE_GROOVY
	GLUE_TYPE__GLUE_SHELL
	GLUE_TYPE__GLUE_PYTHON
	GLUE_TYPE__GLUE_PHP
	GLUE_TYPE__GLUE_NODEJS
	GLUE_TYPE__GLUE_POWERSHELL
)
