package consts

import (
	"github.com/xoctopus/logx"
)

const (
	KEY_TRACE_ID             = "@traceid"
	KEY_TRACE_SPAN_ID        = "@span.id"
	KEY_TRACE_SPAN_NAME      = "@span.name"
	KEY_TRACE_PARENT_SPAN_ID = "@pspan.id"
	KEY_SERVICE_NAME         = "@svc.name"
	KEY_SERVICE_VERSION      = "@svc.version"
	KEY_SOURCE_FUNC          = "@src.func"
	KEY_SOURCE_FILE          = "@src.file"

	KEY_LOG_TIMESTAMP = logx.KEY_TIMESTAMP
	KEY_LOG_LEVEL     = logx.KEY_LEVEL
	KEY_COST          = "@cost"

	KEY_LOG_MESSAGE = logx.KEY_MESSAGE
)

type Format = logx.LogFormat

const (
	JSON = logx.LogFormatJSON
	TEXT = logx.LogFormatTEXT
)

const TIME_FORMAT = logx.TIME_FORMAT
