package appx

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/slicex"
)

// Meta is the build and runtime identity of an application.
type Meta struct {
	Name     string  `json:"name"`
	Feature  string  `json:"feature"`
	Version  string  `json:"version"`
	CommitID string  `json:"commit"`
	Date     string  `json:"date"`
	Runtime  Runtime `json:"runtime"`
}

// DefaultMeta is the baseline [Meta] used when [WithBuildMeta] is omitted.
// Runtime is taken from [GetRuntime] at package init.
var DefaultMeta = Meta{
	Name:     "name",
	Feature:  "branch",
	Version:  "version",
	CommitID: "commit",
	Date:     time.Now().Format("200601021504"),
	Runtime:  GetRuntime(),
}

// String formats Meta as name:feature@version#commit_date(runtime).
func (m *Meta) String() string {
	return fmt.Sprintf("%s:%s@%s#%s_%s(%s)", m.Name, m.Feature, m.Version, m.CommitID, m.Date, m.Runtime)
}

// VersionString joins Name, Version, CommitID, and lowercase Runtime with "-".
func (m *Meta) VersionString() string {
	return strings.Join(slicex.Filter(
		[]string{m.Name, m.Version, m.CommitID, strings.ToLower(string(m.Runtime))},
		func(s string) bool { return len(s) > 0 },
	), "-")
}

// Print writes a colored multi-line Meta summary to stdout.
func (m *Meta) Print() {
	fmt.Printf("%s%s\n", color.HiRedString("name:     "), color.HiYellowString("%s", m.Name))
	fmt.Printf("%s%s\n", color.HiRedString("feature:  "), color.HiYellowString("%s", m.Feature))
	fmt.Printf("%s%s\n", color.HiRedString("version:  "), color.HiYellowString("%s", m.Version))
	fmt.Printf("%s%s\n", color.HiRedString("commit:   "), color.HiYellowString("%s", m.CommitID))
	fmt.Printf("%s%s\n", color.HiRedString("date:     "), color.HiYellowString("%s", m.Date))
	fmt.Printf("%s%s\n", color.HiRedString("runtime:  "), color.HiYellowString("%s", m.Runtime))
	fmt.Printf("\n")
}

// Overwrite copies non-empty fields from meta onto m.
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

// AppOption holds build [Meta] and startup hooks for an [AppCtx].
type AppOption struct {
	*Meta
	// PreRunners run sequentially before Serves: ordered startup init such as
	// global config or context injection.
	PreRunners []func()
	// Serves run in parallel after PreRunners: long-lived services such as
	// HTTP servers or scheduled jobs.
	Serves []func()
	// CloseFns are extra close callbacks for [AppCtx.Close]. Components from
	// Conf are closed automatically and need not be listed here.
	CloseFns []func() error
}

// AppendPreRunners appends ordered startup callbacks.
func (o *AppOption) AppendPreRunners(runners ...func()) {
	o.PreRunners = append(o.PreRunners, runners...)
}

// AppendServes appends long-lived service callbacks.
func (o *AppOption) AppendServes(runners ...func()) {
	o.Serves = append(o.Serves, runners...)
}

// PreRun runs PreRunners then Serves. Invoked by the `run` command.
func (o *AppOption) PreRun() {
	BatchRunSync(o.PreRunners...)
	go BatchRun(o.Serves...)
}

// BatchRun executes runners concurrently and waits for all to finish.
func BatchRun(runners ...func()) {
	wg := &sync.WaitGroup{}
	for i := range runners {
		run := runners[i]

		wg.Go(func() {
			run()
		})
	}
	wg.Wait()
}

// BatchRunSync executes runners sequentially in order.
func BatchRunSync(runners ...func()) {
	for i := range runners {
		runners[i]()
	}
}

type tCtxMeta struct{}

var (
	// AppMetaFrom returns [Meta] from ctx if present.
	// [AppCtx.Conf] injects Meta before types.InitByContext so Initializers
	// (for example confotel) can resolve service name and version.
	AppMetaFrom = contextx.From[tCtxMeta, Meta]
	// MustAppMeta returns [Meta] from ctx or panics when missing.
	MustAppMeta = contextx.Must[tCtxMeta, Meta]
	// WithAppMeta stores [Meta] in ctx.
	WithAppMeta = contextx.With[tCtxMeta, Meta]
	// CarrierAppMeta builds a contextx.Carrier for [Meta].
	CarrierAppMeta = contextx.Carry[tCtxMeta, Meta]
)
