package internal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xoctopus/logx"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/conflogx/internal"
)

func TestNewZapInstance(t *testing.T) {
	t.Run("FixedFileName", func(t *testing.T) {
		dir := t.TempDir()
		logger, stop := internal.NewZapInstance(
			internal.WithOutputDir(dir),
			internal.WithFileName("service.log"),
			internal.WithLevel(logx.LogLevelInfo),
		)
		defer func() { Expect(t, stop(), Succeed()) }()

		logger.Info("fixed file")

		Expect(t, stop(), Succeed())

		data, err := os.ReadFile(filepath.Join(dir, "service.log"))
		Expect(t, err, Succeed())
		Expect(t, string(data), ContainsSubString("fixed file"))
	})

	t.Run("DefaultFileName", func(t *testing.T) {
		dir := t.TempDir()
		logger, stop := internal.NewZapInstance(
			internal.WithOutputDir(dir),
			internal.WithLevel(logx.LogLevelInfo),
		)
		defer func() { Expect(t, stop(), Succeed()) }()

		logger.Info("default file")
		Expect(t, stop(), Succeed())

		data, err := os.ReadFile(filepath.Join(dir, internal.DefaultFileName))
		Expect(t, err, Succeed())
		Expect(t, string(data), ContainsSubString("default file"))
	})

	t.Run("RollingFileName", func(t *testing.T) {
		dir := t.TempDir()
		name := "20260819-15-0.log"

		Expect(t, os.WriteFile(filepath.Join(dir, name), nil, 0o644), Succeed())

		logger, stop := internal.NewZapInstance(
			internal.WithOutputDir(dir),
			internal.WithRolling(true),
			internal.WithLevel(logx.LogLevelInfo),
		)
		defer func() { Expect(t, stop(), Succeed()) }()

		logger.Info("rolling file")
		Expect(t, stop(), Succeed())

		data, err := os.ReadFile(filepath.Join(dir, name))
		Expect(t, err, Succeed())
		Expect(t, string(data), ContainsSubString("rolling file"))
	})

	t.Run("EmptyOutputDirFallsBack", func(t *testing.T) {
		logger, stop := internal.NewZapInstance()
		Expect(t, stop, BeNil[func() error]())
		Expect(t, logger, NotBeNil[logx.Logger]())
		logger.Info("fallback stderr")
	})
}
