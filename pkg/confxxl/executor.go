package confxxl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"cgtech.gitlab.com/saitox/logx"
	"cgtech.gitlab.com/saitox/schex/pkg/synapse"
	"cgtech.gitlab.com/saitox/x/codex"
	"cgtech.gitlab.com/saitox/x/syncx"

	"cgtech.gitlab.com/saitox/confx/pkg/types"
)

func NewExecutor(ctx context.Context, executorName, remoteAddr, listenAddr string, appliers ...ExecutorOptionApplier) (Executor, error) {
	o := ExecutorOption{
		keepalive: 30 * time.Second,
		backlogs:  10,
		client:    http.DefaultClient,
	}
	o.init(executorName, remoteAddr, listenAddr, appliers...)
	e := &executor{ExecutorOption: o}

	e.lastActivity.Store(time.Now().Add(-10 * e.keepalive))
	e.lastReport.Store(time.Now().Add(-10 * e.keepalive))
	e.jobIDs = syncx.NewXmap[int64, string]()
	e.syn = synapse.NewSynapse(
		ctx,
		synapse.WithBeforeCloseFunc(func(ctx context.Context) {
			l := logx.From(ctx).With("executor", e.name)
			_ = e.remove()
			l.Info("executor: unregistered")
			_ = e.jobs.Close()
			l.Info("executor: job manager closed")
		}),
	)
	e.jobs = newJobManager(e.syn.Children())

	if err := e.syn.Spawn(e.heartbeat); err != nil {
		e.syn.Cancel(err)
		return nil, e.syn.Err()
	}
	return e, nil
}

type Executor interface {
	// Register registers a job of `name` with handler `j` to Executor
	Register(name string, j JobHandler) error
	// Kill kills a job by job id
	Kill(int64) error
	// Do execute schedule once with job `name` and trigger data `t`
	Do(t *TriggerRequest) error
	// Close closes executor
	Close() error
	// RefreshActivity refreshes last heartbeat request time from xxl-job
	RefreshActivity()
	// IsActive reports if this Executor is enabled in xxl-job service
	IsActive() bool
	// IsIdle check idle status by job ID.
	IsIdle(int64) bool
}

type executor struct {
	ExecutorOption
	// jobs maintains jobs under executor
	jobs JobManager
	// syn controls executor lifetime
	syn synapse.Synapse
	// lastActivity records the last heartbeat request time from xxl-job
	lastActivity atomic.Value
	// lastReport records the last report request time to xxl-job
	lastReport atomic.Value
	// jobIDs mapping jobID(int64) to task handler key
	jobIDs syncx.Map[int64, string]
}

func (e *executor) RefreshActivity() {
	e.lastActivity.Store(time.Now())
}

func (e *executor) IsActive() bool {
	return time.Since(e.lastActivity.Load().(time.Time)) < 3*e.keepalive &&
		time.Since(e.lastReport.Load().(time.Time)) < 3*e.keepalive
}

func (e *executor) IsIdle(jobID int64) bool {
	if !e.syn.Canceled() {
		if jobName, ok := e.jobIDs.Load(jobID); ok {
			return e.jobs.IsIdle(jobName)
		}
	}
	return false
}

func (e *executor) Kill(jobID int64) error {
	if !e.syn.Canceled() {
		if jobName, ok := e.jobIDs.Load(jobID); ok {
			e.jobIDs.Delete(jobID)
			return e.jobs.Unregister(jobName)
		}
	}
	return nil
}

func (e *executor) Close() error {
	if e.syn != nil {
		e.syn.Cancel(codex.New(ERROR__EXECUTOR_CLOSED))
		return e.syn.Err()
	}
	return nil
}

func (e *executor) Register(name string, h JobHandler) error {
	if e.syn.Canceled() {
		return e.syn.Err()
	}
	return e.jobs.Register(name, h, WithCallback(e.callback), WithMaxPending(e.backlogs))
}

func (e *executor) Do(r *TriggerRequest) (err error) {
	var (
		span    = types.Span()
		ctx     = context.Background()
		cancel  context.CancelFunc
		timeout = time.Duration(r.ExecutorTimeout) * time.Second
		cause   = codex.Errorf(ERROR__EXECUTOR_SCHEDULE_TIMEOUT, "timeout after %ds", r.ExecutorTimeout)
	)

	e.jobIDs.Store(r.JobID, r.ExecutorHandler)
	defer func() {
		cost := span().Milliseconds()
		log := e.logger.With("job", r.ExecutorHandler, "cost", cost)
		if err != nil {
			log.Error(fmt.Errorf("failed to trigger job: %w", err))
		} else {
			log.Info("triggered")
		}
	}()

	if e.syn.Canceled() {
		return e.syn.Err()
	}

	if timeout > 0 {
		ctx, cancel = context.WithTimeoutCause(ctx, timeout, cause)
		defer cancel()
	}

	return e.jobs.Schedule(ctx, r.ExecutorHandler, r)
}

func (e *executor) post(addr string, body []byte) error {
	req, err := http.NewRequest(http.MethodPost, addr, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to new request: %w", err)
	}
	if len(e.token) > 0 {
		req.Header.Add("XXL-JOB-ACCESS-TOKEN", e.token)
	}
	rsp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = rsp.Body.Close() }()
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", rsp.StatusCode)
	}
	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}
	v := &struct {
		Code int `json:"code"`
	}{}
	if err = json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to parse body: %w", err)
	}
	if v.Code != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", rsp.StatusCode)
	}
	return nil
}

func (e *executor) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(e.keepalive)

	defer func() {
		ticker.Stop()
	}()

	do := func() {
		if err := e.register(); err != nil {
			e.logger.Error(fmt.Errorf("heartbeat failed: %w", err))
		}
		e.lastReport.Store(time.Now())
		e.logger.Info("heartbeat done")
	}

	e.logger.Info("heartbeat routine started")
	do()
	for {
		select {
		case <-ticker.C:
			do()
		case <-ctx.Done():
			e.logger.Warn(fmt.Errorf("heartbeat routine stopped: %v", context.Cause(ctx)))
			return
		}
	}
}

func (e *executor) register() error {
	return codex.Wrap(
		ERROR__EXECUTOR_REGISTER_FAILED,
		e.post(e.remote+"/registry", e.report),
	)
}

func (e *executor) remove() error {
	return codex.Wrap(
		ERROR__EXECUTOR_REGISTER_REMOVE_FAILED,
		e.post(e.remote+"/registryRemove", e.report),
	)
}

func (e *executor) callback(r *TriggerRequest, err error) {
	code, msg := http.StatusOK, "success"
	if err != nil {
		code, msg = http.StatusInternalServerError, err.Error()
	}
	body := []byte(fmt.Sprintf(
		`[{"logID":%d,"logDataTim":%d,"handleCode":%d,"handleMsg":"%s"}]`,
		r.LogID, r.LogDateTime, code, msg,
	))
	if _err := e.post(e.remote+"/callback", body); _err != nil {
		e.logger.Error(codex.Wrap(ERROR__EXECUTOR_CALLBACK_FAILED, _err))
	}
}
