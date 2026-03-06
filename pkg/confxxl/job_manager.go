package confxxl

import (
	"context"

	"github.com/xoctopus/logx"
	"github.com/xoctopus/schex/pkg/synapse"
	"github.com/xoctopus/x/codex"
	"github.com/xoctopus/x/syncx"

	"github.com/xoctopus/confx/pkg/confxxl/enums"
)

type JobManager interface {
	// Register registers a job with name, job handler h and option appliers.
	// It starts scheduler of this job.
	Register(name string, h JobHandler, appliers ...JobOptionApplier) error
	// Unregister unregisters job by name and releases scheduler.
	Unregister(name string) error
	// Schedule does once schedule.
	// The result can be retrieved by JobCallback, if it is configurated when Register
	Schedule(context.Context, string, *TriggerRequest) error
	// IsIdle return if job is during scheduling
	IsIdle(name string) bool
	// Close closes all registered Scheduler.
	Close() error
}

func newJobManager(ctx context.Context) JobManager {
	m := &jobManager{
		m: syncx.NewXmap[string, Task](),
	}
	m.syn = synapse.NewSynapse(
		ctx,
		synapse.WithBeforeCloseFunc(func(ctx context.Context) {
			m.m.Range(func(name string, task Task) bool {
				_ = task.Close()
				logx.From(ctx).With("task", name).Info("job manager: task closed")
				return true
			})
		}),
	)

	return m
}

type jobManager struct {
	m   syncx.Map[string, Task]
	syn synapse.Synapse
}

func (x *jobManager) Register(name string, fn JobHandler, appliers ...JobOptionApplier) error {
	if x.syn.Canceled() {
		return codex.New(ERROR__JOB_MANAGER_CLOSED)
	}
	s, err := newScheduler(x.syn.Children(), name, fn, appliers...)
	if err != nil {
		return err
	}
	if _, ok := x.m.LoadOrStore(name, s); ok {
		_ = s.Close()
		return codex.Errorf(ERROR__JOB_MANAGER_REGISTRY_CONFLICT, "name: %s", name)
	}
	return nil
}

func (x *jobManager) Unregister(name string) error {
	if !x.syn.Canceled() {
		j, loaded := x.m.LoadAndDelete(name)
		if loaded {
			return j.Close()
		}
	}
	return nil
}

func (x *jobManager) IsIdle(name string) bool {
	if !x.syn.Canceled() {
		j, ok := x.m.Load(name)
		return ok && j.Pending() == 0
	}
	return false
}

func (x *jobManager) Schedule(ctx context.Context, name string, r *TriggerRequest) error {
	if x.syn.Canceled() {
		return codex.New(ERROR__JOB_MANAGER_CLOSED)
	}

	j, ok := x.m.Load(name)
	if !ok {
		return codex.New(ERROR__JOB_NOT_REGISTERED)
	}

	switch r.ExecutorBlockStrategy {
	case enums.BLOCK_STRATEGY__COVER_EARLY:
		return j.SkipPreviousAndPush(ctx, r)
	case enums.BLOCK_STRATEGY__DISCARD_LATER:
		return j.PushIfIdle(ctx, r)
	case enums.BLOCK_STRATEGY__SERIAL_EXECUTION:
		return j.Push(ctx, r)
	default:
		return j.Push(ctx, r)
	}
}

func (x *jobManager) Close() error {
	if x.syn != nil {
		x.syn.Cancel(codex.New(ERROR__JOB_MANAGER_CLOSED))
		return x.syn.Err()
	}
	return nil
}
