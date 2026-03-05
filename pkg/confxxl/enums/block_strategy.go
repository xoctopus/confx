package enums

// BlockStrategy defines block strategy of xxl-job trigger
// +genx:enum
type BlockStrategy int8

const (
	BLOCK_STRATEGY_UNKNOWN BlockStrategy = iota
	BLOCK_STRATEGY__SERIAL_EXECUTION
	BLOCK_STRATEGY__DISCARD_LATER
	BLOCK_STRATEGY__COVER_EARLY
)
