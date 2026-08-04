// Package appx builds a cobra-backed application context: build metadata,
// env-driven configuration, startup hooks, and shutdown.
//
// Typical flow:
//
//  1. [NewAppContext](Main, options...) with [WithBuildMeta], optional
//     [WithMainRoot] / [WithPreRunner] / [WithServes] / [WithCloseFns].
//  2. [AppCtx.Conf] loads config from the environment (and optional
//     config/local.yml), writes defaults under config/, then initializes
//     fields that implement confx types Init hooks.
//  3. [AppCtx.Execute] dispatches cobra subcommands (`run`, `version`, or
//     commands from [AppCtx.AddCommand]).
//  4. [AppCtx.Close] on shutdown closes Conf components and CloseFns.
//
// See example/appx for an end-to-end binary.
//
// # Meta and context
//
// [Meta] holds Name / Feature / Version / CommitID / Date / Runtime.
// [NewAppContext] starts from [DefaultMeta]; [WithBuildMeta] replaces it.
// [AppCtx.Conf] injects Meta via [WithAppMeta] so nested Initializers
// (e.g. confotel) can read it with [AppMetaFrom] / [MustAppMeta].
//
//	app := appx.NewAppContext(
//		Main,
//		appx.WithBuildMeta(appx.Meta{Name: "demo", Version: "v1.0.0"}),
//		appx.WithMainRoot("."),
//	)
//	app.Conf(ctx, &cfg)
//
// # Configuration
//
// [AppCtx.Conf] accepts one or more config pointers. Each named type becomes
// an envx group (e.g. DEMO__OTEL): merge local.yml → decode env → write
// config/default.yml and config/.env → [types.InitByContext] on fields.
// Anonymous structs are allowed only when a single configuration is passed.
//
// # Runners and shutdown
//
// On `run`, [AppOption.PreRun] runs [WithPreRunner] sequentially, then
// [WithServes] in parallel ([BatchRun]):
//
//   - [WithPreRunner]: ordered startup init (global config, inject context, …)
//   - [WithServes]: long-lived services (HTTP server, cron jobs, …)
//   - [WithCloseFns]: extra close hooks; Conf components close automatically
//     in [AppCtx.Close] and do not need registration
//
// # CLI
//
// [AppCtx.Execute] runs the root cobra command from main.
// [NewAppContext] registers `run` (via [AppCtx.AddCommand]) and `version`.
//
// # Runtime
//
// [GetRuntime] reads [RuntimeEnvKey] (RUNTIME_ENV): PROD / STAGING / DEV
// (default PROD). Stored on [Meta.Runtime].
package appx
