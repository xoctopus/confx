package types

const MaskedPassword = "--------"

// Password as a string underlying and implements SecurityStringer
type Password string

func (p Password) String() string { return string(p) }

func (p Password) SecurityString() string { return MaskedPassword }
