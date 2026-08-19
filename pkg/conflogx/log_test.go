package conflogx_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/xoctopus/logx"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/conflogx"
)

func TestLoggerConfig_WithContext(t *testing.T) {
	t.Run("DefaultSTD", func(t *testing.T) {
		ctx := (&conflogx.LoggerConfig{}).WithContext(context.Background())
		logx.From(ctx).Info("default std")
	})

	t.Run("ZapToFile", func(t *testing.T) {
		dir := t.TempDir()
		cfg := &conflogx.LoggerConfig{
			Level:     logx.LogLevelInfo,
			Format:    logx.LogFormatJSON,
			OutputDIR: dir,
		}
		ctx := cfg.WithContext(context.Background())
		logx.From(ctx).Info("hello file")
		Expect(t, cfg.Close(), Succeed())

		data, err := os.ReadFile(filepath.Join(dir, "app.log"))
		Expect(t, err, Succeed())
		Expect(t, string(data), ContainsSubString("hello file"))
	})

	t.Run("CustomFileName", func(t *testing.T) {
		dir := t.TempDir()
		cfg := &conflogx.LoggerConfig{
			OutputDIR: dir,
			FileName:  "worker.log",
			Rolling:   false,
		}
		ctx := cfg.WithContext(context.Background())
		logx.From(ctx).Info("custom file")
		Expect(t, cfg.Close(), Succeed())

		data, err := os.ReadFile(filepath.Join(dir, "worker.log"))
		Expect(t, err, Succeed())
		Expect(t, string(data), ContainsSubString("custom file"))
	})

	t.Run("RollingFile", func(t *testing.T) {
		dir := t.TempDir()
		cfg := &conflogx.LoggerConfig{
			OutputDIR: dir,
			Rolling:   true,
		}
		ctx := cfg.WithContext(context.Background())
		logx.From(ctx).Info("rolling")
		Expect(t, cfg.Close(), Succeed())

		entries, err := os.ReadDir(dir)
		Expect(t, err, Succeed())
		Expect(t, len(entries), Equal(1))
		Expect(t, entries[0].Name(), MatchRegexp(`^\d{8}-\d{2}-\d+\.log$`))

		data, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
		Expect(t, err, Succeed())
		Expect(t, string(data), ContainsSubString("rolling"))
	})

	t.Run("OutputDIRForcesZap", func(t *testing.T) {
		dir := t.TempDir()
		cfg := &conflogx.LoggerConfig{
			Mode:      conflogx.LogSTD,
			OutputDIR: dir,
		}
		ctx := cfg.WithContext(context.Background())
		logx.From(ctx).Info("forced zap")
		Expect(t, cfg.Close(), Succeed())

		data, err := os.ReadFile(filepath.Join(dir, "app.log"))
		Expect(t, err, Succeed())
		Expect(t, string(data), ContainsSubString("forced zap"))
	})

	t.Run("InvalidOutputDIRFallsBack", func(t *testing.T) {
		cfg := &conflogx.LoggerConfig{
			OutputDIR: filepath.Join("/dev/null", "cannot-create"),
		}

		ctx := cfg.WithContext(context.Background())
		Expect(t, cfg.Close(), Succeed())
		logx.From(ctx).Info("fallback stderr")
	})

	t.Run("CloseFlushesBufferedLogs", func(t *testing.T) {
		dir := t.TempDir()
		cfg := &conflogx.LoggerConfig{OutputDIR: dir}
		ctx := cfg.WithContext(context.Background())
		logx.From(ctx).Info("before close")
		Expect(t, cfg.Close(), Succeed())

		data, err := os.ReadFile(filepath.Join(dir, "app.log"))
		Expect(t, err, Succeed())
		Expect(t, string(data), ContainsSubString("before close"))
	})
}
