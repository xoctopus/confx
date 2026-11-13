package confjwt

// Error confjwt error code
// +genx:code
// @code domain=CONFJWT
type Error int8

const (
	ERROR_UNDEFINED Error = iota
	// ERROR__TOKEN_VALIDATION invalid token
	// @code api=StatusUnauthorized
	ERROR__TOKEN_VALIDATION
	// ERROR__TOKEN_GENERATING failed generating
	ERROR__TOKEN_GENERATING
)
