package conflogx

import (
	"context"
	"strings"
	"time"

	"github.com/xoctopus/logx"

	"github.com/xoctopus/confx/pkg/conflogx/internal"
	"github.com/xoctopus/confx/pkg/types"
)

const (
	LogSTD = "STD"
	LogZap = "ZAP"
)

type LoggerConfig struct {
	Level  logx.LogLevel  `url:",default=debug"`
	Format logx.LogFormat `url:",default=json"`
	Mode   string         `url:",default=std"`

	OutputDIR     string         `url:""`
	FileName      string         `url:""`
	Rolling       bool           `url:""`
	FlushInterval types.Duration `url:",default=30s"`

	stop func() error
}

func (c *LoggerConfig) SetDefault() {
	c.Rolling = true
	if len(c.OutputDIR) == 0 {
		c.OutputDIR = "./log"
	}
	if c.FlushInterval == 0 {
		c.FlushInterval = types.Duration(time.Second * 30)
	}
}

func (c *LoggerConfig) Close() error {
	if c.stop == nil {
		return nil
	}
	stop := c.stop
	c.stop = nil
	return stop()
}

func (c *LoggerConfig) WithContext(ctx context.Context) context.Context {
	logx.SetLogFormat(c.Format)
	logx.SetLogLevel(c.Level)

	mode := strings.ToUpper(c.Mode)
	if c.OutputDIR != "" {
		mode = LogZap
	}

	switch mode {
	case LogZap:
		if c.OutputDIR == "" {
			return logx.With(ctx, logx.NewZap())
		}
		opts := []internal.Option{
			internal.WithOutputDir(c.OutputDIR),
			internal.WithLevel(c.Level),
			internal.WithFlushInterval(time.Duration(c.FlushInterval)),
		}
		if c.Rolling {
			opts = append(opts, internal.WithRolling(true))
		} else if c.FileName != "" {
			opts = append(opts, internal.WithFileName(c.FileName))
		}
		logger, stop := internal.NewZapInstance(opts...)
		c.stop = stop
		return logx.With(ctx, logger)
	default:
		return logx.With(ctx, logx.NewDefault())
	}
}
