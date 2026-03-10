package confrdb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xoctopus/sqlx/pkg/builder"
	"github.com/xoctopus/sqlx/pkg/frag"
	"github.com/xoctopus/sqlx/pkg/migrator"
	"github.com/xoctopus/sqlx/pkg/session"
	"github.com/xoctopus/sqlx/pkg/sql/adaptor"
	_ "github.com/xoctopus/sqlx/pkg/sql/adaptor/mysql"
	"github.com/xoctopus/x/flagx"

	"github.com/xoctopus/confx/pkg/confrdb/option"
	"github.com/xoctopus/confx/pkg/types"
	"github.com/xoctopus/confx/pkg/types/liveness"
)

// endpoint is consistency boundary of a rdb. it is the physical or driver-level
// limit of transactional atomicity. It is defined as the maximum unit for which
// a database can provide atomicity guarantees.
//
// eg:
//
//	a single MySQL instance
//	a standalone PostgreSQL database
//	a serial of database files in a sqlite process
type endpoint[A any] struct {
	types.Endpoint[Option[A]]
	Readonly types.Endpoint[Option[A]]

	database string
	session  string
	catalog  builder.Catalog
	db       adaptor.Adaptor
	ro       adaptor.Adaptor
}

type (
	EndpointMySQL    = endpoint[option.MySQL]
	EndpointPostgres = endpoint[option.Postgres]
	EndpointSQLite   = endpoint[option.SQLite]
)

func (d *endpoint[A]) SetDefault() {
	d.Option.SetDefault()
}

// ApplyCatalog should do before endpoint initialization
func (d *endpoint[A]) ApplyCatalog(catalogs ...builder.Catalog) {
	d.catalog = builder.NewCatalog()

	for _, catalog := range catalogs {
		for table := range catalog.Tables() {
			d.catalog.Add(table)
		}
	}
}

func (d *endpoint[A]) Init(ctx context.Context) error {
	if d.db != nil {
		return nil
	}

	if err := d.Endpoint.Init(); err != nil {
		return fmt.Errorf("failed to init main endpoint: %w", err)
	}
	if x, ok := any(d.Option.AdaptorOption).(option.TLSConfigPatcher); ok && !d.Endpoint.Cert.IsZero() {
		if err := x.WithTLS(d.Endpoint.Cert.Config()); err != nil {
			return fmt.Errorf("failed to patch tls config for adaptor: %w", err)
		}
	}
	if len(d.Option.Name) == 0 && len(d.Endpoint.URL().Path) > 1 {
		name := strings.TrimPrefix(d.Endpoint.URL().Path, "/")
		if parts := strings.Split(name, "/"); len(parts) == 1 {
			d.Option.Name = parts[0]
		} else {
			return fmt.Errorf("invalid dsn or session name: %s", d.Endpoint.SecurityString())
		}
	}

	main := d.Endpoint
	d.database = d.Endpoint.Key()
	d.session = d.Option.Name
	db, err := adaptor.Open(ctx, main.String())
	if err != nil {
		return err
	}

	d.db = db
	d.Option.Apply(d.db.D())

	if !d.Readonly.IsZero() {
		if !d.Readonly.IsZero() {
			if err = d.Readonly.Init(); err != nil {
				return fmt.Errorf("failed to init readonly endpoint: %w", err)
			}
		}
		// readonly endpoint
		ro := d.Readonly
		// reuse main configurations
		if ro.Auth.IsZero() {
			ro.Auth = main.Auth
		}
		ro.AddOption("_ro", "true")
		db, err = adaptor.Open(ctx, ro.String())
		if err != nil {
			return err
		}
		d.ro = db
		d.Option.Apply(d.ro.D())
	}

	session.Register(d.session, d.catalog)

	return d.LivenessCheck(ctx).FailureReason()
}

func (d *endpoint[A]) LivenessCheck(ctx context.Context) (v liveness.Result) {
	v = liveness.NewLivenessData()

	db := d.db
	if d.ro != nil {
		db = d.ro
	}
	if db == nil {
		v.End(errors.New("store session: lost connection"))
		return
	}

	_, err := db.Query(ctx, frag.Query("SELECT 1"))
	v.End(err)
	return
}

// Session is a logical isolation unit and operational handle for database
// adapters.
// eg:
//
//	a specific MySQL database
//	a specific search_path in same PostgreSQL database
//	a particular SQLite database file.
func (d *endpoint[A]) Session() session.Session {
	if d.ro != nil {
		return session.NewReadonly(d.db, d.ro, d.Option.Name)
	}
	return session.New(d.db, d.Option.Name)
}

func (d *endpoint[A]) Catalog() builder.Catalog {
	return d.catalog
}

func (d *endpoint[A]) Run(ctx context.Context) error {
	o := d.Endpoint.Option

	if o.AutoMigration {
		f := flagx.NewFlag[migrator.Mode]()
		if o.DryRun {
			f.With(migrator.DIFF_MODE_DRY_RUN)
		}
		if o.CreateTableOnly {
			f.With(migrator.DIFF_MODE_CREATE_TABLE)
		}
		ctx = migrator.CtxMode.With(ctx, f)
		q, err := migrator.Migrate(ctx, d.db, d.catalog)
		fmt.Println(q)

		if dir, ok := migrator.CtxOutput.From(ctx); ok {
			filename := filepath.Join(dir, d.Option.Name+".sql")
			_ = os.WriteFile(filename, []byte(q), 0o666)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func (d *endpoint[A]) Close() error {
	if d.db != nil {
		if err := d.db.Close(); err != nil {
			return err
		}
	}
	if d.ro != nil {
		if err := d.ro.Close(); err != nil {
			return err
		}
	}
	return nil
}
