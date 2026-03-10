package confrdb_test

import (
	"context"
	"testing"
	"time"

	"github.com/xoctopus/sqlx/pkg/builder"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/hack"
	"github.com/xoctopus/confx/pkg/confrdb/testdata/models"
)

func TestEndpoint(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Minute)
	defer cancel()

	t.Run("Reachable", func(t *testing.T) {
		ep, err := hack.WithMySQL(ctx, t, "mysql://root@localhost:13306/test", models.Catalog)
		Expect(t, err, Succeed())
		s := ep.Session()
		Expect(t, s.Name(), Equal("test"))
		table := s.T(&models.User{})
		Expect(t, table, NotBeNil[builder.Table]())
		Expect(t, table.TableName(), Equal("t_user"))

		Expect(t, ep.Run(ctx), Succeed())
	})

	t.Run("Unreachable", func(t *testing.T) {
		_, err := hack.WithMySQL(ctx, t, "mysql://root@localhost:3308/test")
		Expect(t, err, Failed())
	})
}
