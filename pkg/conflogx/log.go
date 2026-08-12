package conflogx

import (
	"context"
	"strings"

	"github.com/xoctopus/logx"
)

const (
	LogSTD = "STD"
	LogZap = "ZAP"
)

type LoggerConfig struct {
	Level  logx.LogLevel  `url:",default=debug"`
	Format logx.LogFormat `url:",default=json"`
	Mode   string         `url:",default=std"`
}

func (c *LoggerConfig) WithContext(ctx context.Context) context.Context {
	logx.SetLogFormat(c.Format)
	logx.SetLogLevel(c.Level)

	switch strings.ToUpper(c.Mode) {
	case LogZap:
		return logx.With(ctx, logx.NewZap())
	default:
		return logx.With(ctx, logx.NewDefault())
	}
}
