package main

import (
	"context"
	"os"
	"path/filepath"

	_ "github.com/xoctopus/genx/devpkg"
	"github.com/xoctopus/genx/pkg/genx"
	_ "github.com/xoctopus/sqlx/devpkg"
	"github.com/xoctopus/x/misc/must"
)

func main() {
	cwd := must.NoErrorV(os.Getwd())

	ctx := genx.NewContext(&genx.Args{
		Entrypoint: []string{
			filepath.Join(cwd, "pkg", "confrdb", "testdata", "models"),
		},
	})

	if err := ctx.Execute(context.Background(), genx.Get()...); err != nil {
		panic(err)
	}
}
