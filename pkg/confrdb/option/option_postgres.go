package option

import (
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Postgres struct {
}

func (o *Postgres) SetDefault() {
	panic("not implemented")
}
