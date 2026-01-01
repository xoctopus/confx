package main

import (
	"context"
	"os"
	"path/filepath"

	_ "github.com/xoctopus/genx/devpkg/docx"
	"github.com/xoctopus/genx/pkg/genx"
	"github.com/xoctopus/x/misc/must"
)

func main() {
	cwd := must.NoErrorV(os.Getwd())

	ctx := genx.NewContext(&genx.Args{
		Entrypoint: []string{
			filepath.Join(cwd, "example", "cmdx", "pkg", "..."),
		},
	})

	if err := ctx.Execute(context.Background(), genx.Get()...); err != nil {
		panic(err)
	}
}
