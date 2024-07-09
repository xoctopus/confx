package confapp

import (
	"fmt"
	"sync"
	"time"

	"github.com/fatih/color"
)

type Meta struct {
	Name     string  `json:"name"`
	Feature  string  `json:"feature"`
	Version  string  `json:"version"`
	CommitID string  `json:"commit"`
	Date     string  `json:"date"`
	Runtime  Runtime `json:"runtime"`
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

func (m *Meta) Print() {
	fmt.Printf("%s%s\n", color.HiRedString("name:     "), color.HiYellowString("%s", m.Name))
	fmt.Printf("%s%s\n", color.HiRedString("feature:  "), color.HiYellowString("%s", m.Feature))
	fmt.Printf("%s%s\n", color.HiRedString("version:  "), color.HiYellowString("%s", m.Version))
	fmt.Printf("%s%s\n", color.HiRedString("commit:   "), color.HiYellowString("%s", m.CommitID))
	fmt.Printf("%s%s\n", color.HiRedString("date:     "), color.HiYellowString("%s", m.Date))
	fmt.Printf("%s%s\n", color.HiRedString("runtime:  "), color.HiYellowString("%s", m.Runtime))
	fmt.Printf("\n")
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
	GenMakefile   bool
	GenDockerfile bool

	// PreRunner must run before main
	PreRunners []func()
	// BatchRunner routines need pre-run before enter main, e.g. modules initializations
	BatchRunners []func()
}

func (o *AppOption) NeedAttach() bool {
	return o.GenMakefile || o.GenDockerfile
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
