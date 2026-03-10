package option

import (
	_ "modernc.org/sqlite"
)

type SQLite struct {
}

func (o *SQLite) SetDefault() {
	panic("not implemented")
}
