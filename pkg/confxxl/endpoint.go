package confxxl

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"cgtech.gitlab.com/saitox/logx"
	"cgtech.gitlab.com/saitox/schex/pkg/synapse"
	"cgtech.gitlab.com/saitox/x/codex"
	"cgtech.gitlab.com/saitox/x/urlx"

	"cgtech.gitlab.com/saitox/confx/pkg/helper"
	"cgtech.gitlab.com/saitox/confx/pkg/types"
)

type Endpoint struct {
	types.Endpoint[Option]

	client *http.Client
	server *http.Server

	// listen url info for accepting action from xxl-job
	listen *urlx.URL
	// remote url info for reporting to xxl-job
	remote *urlx.URL
	// executors management. readonly after initialized
	executors map[string]Executor
	// syn controls Endpoint lifetime
	syn synapse.Synapse
}

func (e *Endpoint) SetDefault() {
	e.Option.SetDefault()
}

func (e *Endpoint) Init(ctx context.Context) (err error) {
	e.remote, err = urlx.Parse(
		e.Address,
		urlx.WithPath("/xxl-job-admin/api"),
		urlx.TrimQuery(),
		urlx.TrimFragment(),
	)
	if err != nil {
		return err
	}

	e.listen, err = urlx.Parse(
		e.Option.Listener,
		urlx.WithPath(helper.HostIdentifier(e.Option.ClientID)),
		urlx.TrimQuery(),
		urlx.TrimFragment(),
	)
	if err != nil {
		return err
	}

	e.client = &http.Client{Timeout: time.Duration(e.Option.ClientTimeout)}
	e.executors = make(map[string]Executor)
	e.syn = synapse.NewSynapse(
		ctx,
		synapse.WithBeforeCloseFunc(func(ctx context.Context) {
			if e.server != nil {
				_ = e.server.Shutdown(ctx)
			}
			for _, ex := range e.executors {
				_ = ex.Close()
			}
			logx.From(ctx).Info("endpoint: executors closed")
		}),
	)

	defer func() {
		if err != nil {
			e.syn.Cancel(err)
			err = e.syn.Err()
		}
	}()

	for _, name := range e.Option.Executor {
		if _, ok := e.executors[name]; ok {
			continue
		}
		var ex Executor
		listen := urlx.From(e.listen.URL, urlx.WithPath(e.listen.Path+"/"+name))
		ex, err = NewExecutor(
			e.syn.Children(), name, e.remote.String(), listen.String(),
			WithHttpClient(e.client),
			WithAccessToken(e.Option.AccessToken),
			WithLogger(logx.From(ctx).With("executor", name)),
			WithKeepAliveInterval(time.Duration(e.Option.KeepAlive)),
			WithMaxBacklogs(e.Option.MaxBacklogs),
		)
		if err != nil {
			return err
		}
		e.executors[name] = ex
	}

	return e.syn.Spawn(e.serve)
}

func (e *Endpoint) Close() error {
	if e.syn != nil {
		e.syn.Cancel(codex.New(ERROR__ENDPOINT_CLOSED))
		return e.syn.Err()
	}
	return nil
}

func (e *Endpoint) RegisterJob(exec string, job string, fn JobHandler) error {
	if e.syn.Canceled() {
		return e.syn.Err()
	}
	ex, ok := e.executors[exec]
	if !ok {
		return codex.New(ERROR__EXECUTOR_NOT_REGISTERED)
	}
	return ex.Register(job, fn)
}

func (e *Endpoint) IsActive(exec string) bool {
	if e.syn.Canceled() {
		return false
	}
	ex, ok := e.executors[exec]
	if !ok {
		return false
	}
	return ex.IsActive()
}

func (e *Endpoint) serve(ctx context.Context) {
	var err error

	// pattern served http handler pattern for accepting triggers from xxl-job
	// :listen.Port/{host-identifier}/{executor}/{job-name}/{xxl-actions}
	pattern := fmt.Sprintf("%s/{%s}/{%s}", e.listen.Path, PATH_EXECUTOR, PATH_ACTION)

	log := logx.From(ctx).With("port", e.listen.Port(), "api", pattern)
	log.Info("xxl-job serve started")
	defer func() {
		if err != nil {
			log.Error(fmt.Errorf("xxl-job serve closed cause by: %w", err))
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("POST "+pattern, e.handle)

	e.server = &http.Server{
		Addr:    ":" + strconv.Itoa(int(e.listen.Port())),
		Handler: mux,
	}

	if !e.Cert.IsZero() {
		e.server.TLSConfig = e.Cert.Config()
	}
	err = e.server.ListenAndServe()
}

func (e *Endpoint) handle(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		name   = r.PathValue(PATH_EXECUTOR)
		action = r.PathValue(PATH_ACTION)
		log    = logx.From(e.syn.Children()).With(PATH_EXECUTOR, name, PATH_ACTION, action)
	)

	defer func() {
		w.Header().Add("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"code":500,"msg","%v"}`, err)))
			log.Error(err)
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"code":200,"msg","success"}`))
			log.Info("handled")
		}
	}()

	ex, ok := e.executors[name]
	if !ok {
		err = fmt.Errorf("executor not found: %s", name)
		return
	}
	// each request from xxl-job should refresh last activity
	ex.RefreshActivity()

	switch action {
	case "beat":
		return
	case "idleBeat":
		input := &IdleBeatRequest{}
		if err = bind(r, input); err != nil {
			err = fmt.Errorf("failed to bind request body: %w", err)
			return
		}
		log = log.With("job_id", input.JobID)
		if !ex.IsIdle(input.JobID) {
			err = fmt.Errorf("job busy")
		}
		return
	case "kill":
		input := &KillJobRequest{}
		if err = bind(r, input); err != nil {
			err = fmt.Errorf("failed to bind request body: %w", err)
			return
		}
		log = log.With("job_id", input.JobID)
		err = ex.Kill(input.JobID)
		return
	case "run":
		input := &TriggerRequest{}
		if err = bind(r, input); err != nil {
			err = fmt.Errorf("failed to bind request body: %w", err)
			return
		}
		log = log.With("job_id", input.JobID, "job_name", input.ExecutorHandler)
		err = ex.Do(input)
		return
	case "log":
		// security considerations:
		// 1. prevent leakage of sensitive data.
		// 2. ensure the storage integrity of XXL-JOB logs.
		// performance Considerations:
		// 1. xxl-job's log storage is overly simplistic (mysql-based), which
		//    creates database pressure and requires periodic cleanup.
		// 2. high network overhead may lead to frequent "misfire" incidents.
		return
	default:
		err = codex.Errorf(ERROR__XXL_INVALID_ACTION, "got action: %s", action)
	}
}
