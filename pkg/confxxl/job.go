package confxxl

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/xoctopus/schex/pkg/schex"
	"github.com/xoctopus/schex/pkg/synapse"
	"github.com/xoctopus/x/codex"
)

type Task interface {
	// Name denotes Task name
	Name() string
	// Close stops job scheduling
	Close() error
	// Push appends a task for scheduling
	Push(context.Context, *TriggerRequest) error
	// Pending returns pending task count
	Pending() int
	// Scheduling returns scheduling task count
	Scheduling() int
	// SkipPreviousAndPush appending task after skip previous tasks
	SkipPreviousAndPush(ctx context.Context, r *TriggerRequest) error
	// PushIfIdle appending task if current scheduler is idle
	PushIfIdle(ctx context.Context, r *TriggerRequest) error
}

func newScheduler(ctx context.Context, name string, fn JobHandler, appliers ...JobOptionApplier) (Task, error) {
	o := &JobOption{}
	for _, applier := range appliers {
		applier(o)
	}

	j := &task{
		name:     name,
		backlogs: o.backlogs,
	}
	j.fn = schex.JobFunc[*TriggerRequest](func(ctx context.Context, r *TriggerRequest) error {
		j.scheduling.Add(1)
		defer j.scheduling.Add(-1)
		return fn(ctx, r)
	})

	if o.cb != nil {
		j.cb = schex.HandlerCallback[*TriggerRequest](o.cb)
	}
	if j.backlogs == 0 {
		j.backlogs = 1
	}
	if err := j.skip(ctx); err != nil {
		return nil, err
	}
	return j, nil
}

type task struct {
	name       string
	fn         schex.Job[*TriggerRequest]
	cb         schex.HandlerCallback[*TriggerRequest]
	backlogs   int
	scheduling atomic.Int32

	// mtx keep syn(switch when skipping) and sche in critical zone
	mtx  sync.Mutex
	syn  synapse.Synapse
	sche schex.Scheduler[*TriggerRequest]
}

func (j *task) skip(ctx context.Context) error {
	if j.syn != nil {
		if j.sche.Pending() == 0 {
			return nil
		}
		j.syn.Cancel(codex.New(ERROR__JOB_PENDING_TASK_SKIP))
		<-j.syn.Done()
	}

	j.syn = synapse.NewSynapse(
		ctx,
		synapse.WithBeforeCloseFunc(func(ctx context.Context) {
			if j.sche != nil {
				_ = j.sche.Close()
			}
		}),
	)
	j.sche = schex.NewScheduler(
		j.fn,
		schex.WithParallel[*TriggerRequest](1),
		schex.WithMaxPending[*TriggerRequest](j.backlogs),
		schex.WithCallback[*TriggerRequest](j.cb),
		schex.WithExitCallback[*TriggerRequest](func(pending []*TriggerRequest, err error) {
			for _, r := range pending {
				if j.cb != nil {
					j.cb(r, err)
				}
			}
		}),
	)
	if err := j.sche.Run(j.syn.Children()); err != nil {
		j.syn.Cancel(err)
		return j.syn.Err()
	}
	return nil
}

func (j *task) Name() string {
	return j.name
}

func (j *task) Close() error {
	j.mtx.Lock()
	defer j.mtx.Unlock()
	j.syn.Cancel(codex.New(ERROR__JOB_MANUAL_CLOSE))
	return j.syn.Err()
}

func (j *task) Push(ctx context.Context, r *TriggerRequest) error {
	j.mtx.Lock()
	defer j.mtx.Unlock()
	if !j.syn.Canceled() {
		return j.sche.Push(ctx, r)
	}
	return nil
}

func (j *task) PushIfIdle(ctx context.Context, r *TriggerRequest) error {
	j.mtx.Lock()
	defer j.mtx.Unlock()
	if !j.syn.Canceled() {
		if j.Pending()+j.Scheduling() == 0 {
			return j.sche.Push(ctx, r)
		}
		return codex.New(ERROR__JOB_SCHEDULER_BUSY)
	}
	return codex.New(ERROR__JOB_CLOSED)
}

func (j *task) SkipPreviousAndPush(ctx context.Context, r *TriggerRequest) error {
	j.mtx.Lock()
	defer j.mtx.Unlock()
	if !j.syn.Canceled() {
		if err := j.skip(j.syn.Parent()); err != nil {
			return err
		}
		return j.sche.Push(ctx, r)
	}
	return codex.New(ERROR__JOB_CLOSED)
}

func (j *task) Pending() int {
	j.mtx.Lock()
	defer j.mtx.Unlock()
	return j.sche.Pending()
}

func (j *task) Scheduling() int {
	return int(j.scheduling.Load())
}
