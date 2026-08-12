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

// DefaultMeta is the baseline [Meta] used when [WithMeta] is omitted.
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
	// PreInits run sequentially before component Init in [AppCtx.Conf]: prepare
	// metadata or hooks so Init can succeed, without opening business traffic.
	PreInits []func()
	// PreRunners run sequentially before Serve: ordered process-level prep
	// with a complete ctx, such as global config or context injection.
	// Registered via [WithPreRun].
	PreRunners []func()
	// Serves run in parallel after PreRun: long-lived services such as
	// HTTP servers or scheduled jobs. Registered via [WithServe].
	Serves []func()
	// CloseFns are extra close callbacks for [AppCtx.Close], registered via
	// [WithClose]. Components from Conf are closed automatically and need not
	// be listed here.
	CloseFns []func() error
}

// PreInit runs [WithPreInit] callbacks sequentially before component Init.
func (o *AppOption) PreInit() {
	BatchRunSync(o.PreInits...)
}

// PreRun runs [WithPreRun] callbacks sequentially. Invoked by `run`
// before [AppOption.Serve].
func (o *AppOption) PreRun() {
	BatchRunSync(o.PreRunners...)
}

// Serve starts [WithServe] callbacks in parallel (non-blocking). Invoked by
// `run` after [AppOption.PreRun].
func (o *AppOption) Serve() {
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
