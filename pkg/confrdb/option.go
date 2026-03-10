package confrdb

import (
	"database/sql"
	"time"

	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/textx"

	"github.com/xoctopus/confx/pkg/confrdb/option"
	"github.com/xoctopus/confx/pkg/types"
)

type Option[A any] struct {
	// AutoMigration if enable auto migration
	AutoMigration bool `url:"-"`
	// DryRun migration with dry run mode
	DryRun bool `url:"-"`
	// CreateTableOnly just do table creations without column diff and modifications
	CreateTableOnly bool `url:"-"`

	// MaxOpenConns the upper limit on open connections. this should be tuned
	// based on both application concurrency and database server capacity.
	// a typical recommendation is 2-4 times the number of CPU cores of database
	// sever.
	MaxOpenConns int `url:"-"`
	// MaxIdleConns is recommended to be set to the same value as MaxOpenConns.
	MaxIdleConns int `url:"-"`
	// ConnMaxLifetime is recommended to be 0.5 to 1 hour.
	// NOTE: it should lower than sever-side `wait_timeout` setting. (eg: MySQL's `wait_timeout`) .
	// server-side controls the maximum duration connection remains open while idle.
	ConnMaxLifetime types.Duration `url:"-"`
	// ConnMaxIdleTime is recommended to be 5 to 10 minutes to release rapidly of
	// reclaim connection resources.
	ConnMaxIdleTime types.Duration `url:"-"`

	// Name denotes a globally **unique** identifier for a database endpoint and
	// its lifecycle sessions.
	Name string `url:"-"`
	// AdaptorOption for different driver. eg mysql, postgres etc.
	AdaptorOption A `url:",inline"`
}

func (o *Option[A]) SetDefault() {
	if x, ok := any(&o.AdaptorOption).(interface{ SetDefault() }); ok {
		x.SetDefault()
	}

	filebased := false
	switch any(o.AdaptorOption).(type) {
	// TODO for different adaptor
	// case option.Postgres:
	// 	if len(o.Name) == 0 {
	// 		o.Name = "public"
	// 	}
	// case option.SQLite:
	// 	if len(o.Name) == 0 {
	// 		o.Name = "main"
	// 	}
	// 	// _joural=WAL may set max open to runtime.NumCPU
	// 	o.MaxOpenConns = 1
	// 	o.MaxIdleConns = 1
	// 	o.ConnMaxLifetime = 0
	// 	o.ConnMaxIdleTime = 0
	// 	sqlite = true
	case option.SQLite, option.DuckDB:
		filebased = true
	}

	if o.ConnMaxLifetime == 0 && !filebased {
		o.ConnMaxLifetime = types.Duration(time.Hour / 2)
	}
	if o.ConnMaxIdleTime == 0 && !filebased {
		o.ConnMaxIdleTime = types.Duration(time.Minute * 10)
	}
	if o.MaxOpenConns == 0 {
		o.MaxOpenConns = 100
	}
	if o.MaxIdleConns == 0 {
		o.MaxIdleConns = o.MaxOpenConns
	}

	must.NoErrorV(textx.SetDefault(o))
}

func (o *Option[A]) Apply(db *sql.DB) {
	db.SetMaxOpenConns(o.MaxOpenConns)
	db.SetMaxIdleConns(o.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(o.ConnMaxLifetime))
	db.SetConnMaxIdleTime(time.Duration(o.ConnMaxIdleTime))
}
