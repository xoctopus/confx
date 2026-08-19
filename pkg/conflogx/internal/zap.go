package internal

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/xoctopus/logx"
	"go.uber.org/zap/zapcore"
)

const (
	DefaultFileName        = "app.log"
	defaultBufferSize      = 256 * 1024
	defaultBufferFlushWait = 30 * time.Second
)

type config struct {
	dir           string
	fileName      string
	rolling       bool
	level         logx.LogLevel
	flushInterval time.Duration
}

type Option func(*config)

func WithOutputDir(dir string) Option {
	return func(c *config) {
		c.dir = dir
	}
}

func WithFileName(name string) Option {
	return func(c *config) {
		c.fileName = name
	}
}

func WithRolling(enabled bool) Option {
	return func(c *config) {
		c.rolling = enabled
	}
}

func WithLevel(level logx.LogLevel) Option {
	return func(c *config) {
		c.level = level
	}
}

func WithFlushInterval(d time.Duration) Option {
	return func(c *config) {
		c.flushInterval = d
	}
}

func NewZapInstance(opts ...Option) (logx.Logger, func() error) {
	cfg := config{
		fileName: DefaultFileName,
		level:    logx.LogLevelDebug,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	w, err := openWriter(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "conflogx: failed to open log file in %s: %v\n", cfg.dir, err)
		return logx.NewZap(), nil
	}

	flushInterval := cfg.flushInterval
	if flushInterval <= 0 {
		flushInterval = defaultBufferFlushWait
	}

	bws := &zapcore.BufferedWriteSyncer{
		WS:            zapcore.AddSync(w),
		Size:          defaultBufferSize,
		FlushInterval: flushInterval,
	}
	return logx.NewWithInstance(logx.ZapLogger(bws, 2, cfg.level)), bws.Stop
}

func openWriter(cfg config) (io.Writer, error) {
	if cfg.dir == "" {
		return nil, fmt.Errorf("output dir is empty")
	}
	if cfg.rolling {
		return openRollingWriter(cfg.dir)
	}
	if cfg.fileName == "" {
		cfg.fileName = DefaultFileName
	}
	return openFixedWriter(cfg.dir, cfg.fileName)
}
