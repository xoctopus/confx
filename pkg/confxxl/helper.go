package confxxl

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/xoctopus/confx/pkg/confxxl/enums"
)

type TriggerRequest struct {
	// JobID 任务ID
	JobID int64 `json:"jobId"`
	// ExecutorHandler 任务标识
	ExecutorHandler string `json:"executorHandler"`
	// ExecutorParams 任务参数
	ExecutorParams string `json:"executorParams"`
	// ExecutorBlockStrategy 任务阻塞策略
	ExecutorBlockStrategy enums.BlockStrategy `json:"executorBlockStrategy"`
	// ExecutorTimeout 任务超时时间(秒,大于零生效)
	ExecutorTimeout int64 `json:"executorTimeout"`
	// LogID 本次调度日志ID
	LogID int64 `json:"logId"`
	// LogDateTime 本次调度日志时间
	LogDateTime int64 `json:"logDateTime"`
	// GlueType 任务模式
	GlueType enums.GlueType `json:"glueType"`
	// GlueSource GLUE脚本代码
	GlueSource string `json:"glueSource"`
	// GlueUpdateTime GLUE脚本更新时间，用于判定脚本是否变更以及是否需要刷新
	GlueUpdateTime int64 `json:"glueUpdatetime"`
	// BroadcastIndex 当前分片
	BroadcastIndex int64 `json:"broadcastIndex"`
	// BroadcastTotal 总分片
	BroadcastTotal int64 `json:"broadcastTotal"`
}

type IdleBeatRequest struct {
	JobID int64 `json:"jobID"`
}

type KillJobRequest struct {
	JobID int64 `json:"jobID"`
}

func bind(r *http.Request, v any) error {
	defer func() { _ = r.Body.Close() }()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, v); err != nil {
		return err
	}
	return err
}

// these consts defines exposed api parameters key in path
// eg: /{host-unique-identifier}/{executor}/{job_name}/{action}
const (
	PATH_EXECUTOR = "executor"
	PATH_JOB_NAME = "job_name"
	PATH_ACTION   = "action"
)
