package internal_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

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
		logger, stop := internal.NewZapInstance(
			internal.WithOutputDir(dir),
			internal.WithRolling(true),
			internal.WithLevel(logx.LogLevelInfo),
		)
		defer func() { Expect(t, stop(), Succeed()) }()

		logger.Info("rolling file")
		Expect(t, stop(), Succeed())

		entries, err := os.ReadDir(dir)
		Expect(t, err, Succeed())
		Expect(t, len(entries), Equal(1))
		Expect(t, entries[0].Name(), MatchRegexp(`^\d{8}-\d{2}-\d+\.log$`))

		data, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
		Expect(t, err, Succeed())
		Expect(t, string(data), ContainsSubString("rolling file"))
	})

	t.Run("EmptyOutputDirFallsBack", func(t *testing.T) {
		logger, stop := internal.NewZapInstance()
		Expect(t, stop, BeNil[func() error]())
		Expect(t, logger, NotBeNil[logx.Logger]())
		logger.Info("fallback stderr")
	})

	t.Run("FlushInterval", func(t *testing.T) {
		dir := t.TempDir()
		logger, stop := internal.NewZapInstance(
			internal.WithOutputDir(dir),
			internal.WithLevel(logx.LogLevelInfo),
			internal.WithFlushInterval(50*time.Millisecond),
		)
		defer func() { Expect(t, stop(), Succeed()) }()

		logger.Info("periodic flush")
		time.Sleep(100 * time.Millisecond)

		data, err := os.ReadFile(filepath.Join(dir, internal.DefaultFileName))
		Expect(t, err, Succeed())
		Expect(t, string(data), ContainsSubString("periodic flush"))
	})
}
