package hack

import (
	"context"
	"testing"

	"github.com/xoctopus/sqlx/pkg/builder"

	"github.com/xoctopus/confx/pkg/confrdb"
	"github.com/xoctopus/confx/pkg/confrdb/option"
	"github.com/xoctopus/confx/pkg/types"
)

func WithMySQL(ctx context.Context, t testing.TB, dsn string, catalogs ...builder.Catalog) (*confrdb.EndpointMySQL, error) {
	ep := &confrdb.EndpointMySQL{
		Endpoint: types.Endpoint[confrdb.Option[option.MySQL]]{
			Address: dsn,
			Option: confrdb.Option[option.MySQL]{
				AutoMigration: true,
			},
		},
	}
	ep.SetDefault()
	ep.ApplyCatalog(catalogs...)

	err := retrier.Do(func() error {
		return ep.Init(ctx)
	})
	if err != nil {
		return nil, err
	}

	t.Cleanup(func() {
		_ = ep.Close()
	})
	return ep, nil
}
