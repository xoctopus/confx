package confxxl

import (
	"fmt"
	"net/http"
	"time"

	"cgtech.gitlab.com/saitox/logx"
	"cgtech.gitlab.com/saitox/x/misc/must"
	"cgtech.gitlab.com/saitox/x/textx"

	"cgtech.gitlab.com/saitox/confx/pkg/types"
)

type Option struct {
	ClientID string
	// Listener denotes local exposed address. this address reports to xxl-job
	// service for accepting trigger request from xxl-job
	Listener string `url:""`
	// AccessToken xxl-job request access token
	AccessToken string `url:""`
	// KeepAlive heartbeat interval for executors
	KeepAlive types.Duration `url:",default=30s"`
	// ClientTimeout heartbeat interval for executors
	ClientTimeout types.Duration `url:",default=5s"`
	// DisableDebugLog if disable debug log
	DisableDebugLog bool `url:",default=false"`
	// MaxBacklogs denotes max backlogs per job. if job trigger parameter use
	// enums.BLOCK_STRATEGY__SERIAL_EXECUTION mode it may cause backlogs.
	// when backlogs reached this limitation the trigger will be callback with
	// ERROR__JOB_REACH_MAX_BACKLOG error
	MaxBacklogs int `url:",default=10"`
	// Executor registered executor list
	Executor []string
}

func (o *Option) SetDefault() {
	must.NoErrorV(textx.SetDefault(o))
}

type JobOption struct {
	cb       JobCallback
	backlogs int
}

type JobOptionApplier func(*JobOption)

func WithCallback(cb JobCallback) JobOptionApplier {
	return func(o *JobOption) {
		o.cb = cb
	}
}

func WithMaxPending(n int) JobOptionApplier {
	return func(o *JobOption) {
		o.backlogs = n
	}
}

type ExecutorOptionApplier func(*ExecutorOption)

type ExecutorOption struct {
	// name of executor
	name string
	// logger of executor
	logger logx.Logger
	// keepalive denotes heartbeat interval
	keepalive time.Duration
	// backlogs the max backlogged tasks in each task scheduler
	backlogs int
	// token xxl-job service access token
	token string
	// remote xxl-job service api address
	remote string
	// listen denotes http serving address for xxl-job service triggering
	listen string
	// client http client of executor
	client *http.Client
	// report is invariable body when reporting registry to xxl-job service.
	report []byte
}

func (o *ExecutorOption) init(name, remote, listen string, appliers ...ExecutorOptionApplier) {
	must.BeTrueF(len(remote) > 0, "missing required remote address.")
	must.BeTrueF(len(listen) > 0, "missing required listen address.")

	o.name = name
	o.remote = remote
	o.listen = listen
	for _, applier := range appliers {
		applier(o)
	}

	o.report = []byte(fmt.Sprintf(
		`{"registryGroup":"EXECUTOR","registryKey":"%s","registryValue":"%s"}`,
		o.name, o.listen,
	))
	if o.logger == nil {
		o.logger = logx.NewDefault().With("executor", o.name)
	}
	if o.client == nil {
		o.client = http.DefaultClient
	}
}

func WithHttpClient(client *http.Client) ExecutorOptionApplier {
	return func(e *ExecutorOption) {
		e.client = client
	}
}

func WithLogger(logger logx.Logger) ExecutorOptionApplier {
	return func(e *ExecutorOption) {
		e.logger = logger
	}
}

func WithAccessToken(token string) ExecutorOptionApplier {
	return func(e *ExecutorOption) {
		e.token = token
	}
}

func WithKeepAliveInterval(keepalive time.Duration) ExecutorOptionApplier {
	return func(e *ExecutorOption) {
		e.keepalive = keepalive
	}
}

func WithMaxBacklogs(backlogs int) ExecutorOptionApplier {
	return func(e *ExecutorOption) {
		e.backlogs = backlogs
	}
}
