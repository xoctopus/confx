package confapp

import (
	"fmt"
	"sync"
	"time"
)

type Meta struct {
	Name     string
	Feature  string
	Version  string
	CommitID string
	Date     string
	Runtime  Runtime
}

var DefaultMeta = Meta{
	Name:     "name",
	Feature:  "main",
	Version:  "v0.0.0",
	CommitID: "commit",
	Date:     time.Now().Format("200601021504"),
	Runtime:  GetRuntime(),
}

func (m *Meta) String() string {
	return fmt.Sprintf("%s:%s@%s#%s_%s(%s)", m.Name, m.Feature, m.Version, m.CommitID, m.Date, m.Runtime)
}

func (m *Meta) Overwrite(meta Meta) {
	if meta.Name != "" {
		m.Name = meta.Name
	}
	if meta.Feature != "" {
		m.Feature = meta.Feature
	}
	if meta.Version != "" {
		m.Version = meta.Version
	}
	if meta.CommitID != "" {
		m.CommitID = meta.CommitID
	}
	if meta.Date != "" {
		m.Date = meta.Date
	}
	if meta.Runtime != "" {
		m.Runtime = meta.Runtime
	}
}

type AppOption struct {
	Meta
	GenDefaults   bool
	GenMakefile   bool
	GenDockerfile bool

	// PreRunner must run before main
	PreRunners []func()
	// BatchRunner routines need pre-run before enter main, e.g. modules initializations
	BatchRunners []func()
}

func (o *AppOption) NeedAttach() bool {
	return o.GenDefaults || o.GenMakefile || o.GenDockerfile
}

func (o *AppOption) AppendPreRunners(runners ...func()) {
	o.PreRunners = append(o.PreRunners, runners...)
}

func (o *AppOption) AppendBatchRunners(runners ...func()) {
	o.BatchRunners = append(o.BatchRunners, runners...)
}

func (o *AppOption) PreRun() {
	BatchRun(o.PreRunners...)
	go BatchRun(o.BatchRunners...)
}

func BatchRun(runners ...func()) {
	wg := &sync.WaitGroup{}
	for i := range runners {
		run := runners[i]
		wg.Add(1)

		go func() {
			defer wg.Done()
			run()
		}()
	}
	wg.Wait()
}
