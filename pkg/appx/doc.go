// Package appx builds a cobra-backed application context: build metadata,
// env-driven configuration, startup hooks, and shutdown.
//
// Typical flow:
//
//  1. [NewAppContext](Main, options...) with [WithMeta], optional
//     [WithRoot] / [WithPreInit] / [WithPreRun] / [WithServe] /
//     [WithClose].
//  2. [AppCtx.Conf] loads config from the environment (and optional
//     config/local.yml), writes defaults under config/, runs [WithPreInit],
//     then initializes fields that implement confx types Init hooks.
//  3. [AppCtx.Execute] dispatches cobra subcommands (`run`, `version`, or
//     commands from [AppCtx.AddCommand]).
//  4. [AppCtx.Close] on shutdown closes Conf components and [WithClose] hooks.
//
// See example/appx for an end-to-end binary.
//
// # Meta and context
//
// [Meta] holds Name / Feature / Version / CommitID / Date / Runtime.
// [NewAppContext] starts from [DefaultMeta]; [WithMeta] replaces it.
// [AppCtx.Conf] injects Meta via [WithAppMeta] so nested Initializers
// (e.g. confotel) can read it with [AppMetaFrom] / [MustAppMeta].
//
//	app := appx.NewAppContext(
//		Main,
//		appx.WithMeta(appx.Meta{Name: "demo", Version: "v1.0.0"}),
//		appx.WithRoot("."),
//	)
//	app.Conf(ctx, &cfg)
//
// # Configuration
//
// [AppCtx.Conf] accepts one or more config pointers. Each named type becomes
// an envx group (e.g. DEMO__OTEL): merge local.yml → decode env → write
// config/default.yml and config/.env → [WithPreInit] → [types.InitByContext]
// on fields.
// Anonymous structs are allowed only when a single configuration is passed.
//
// # Lifecycle hooks and shutdown
//
// Duty boundary:
//
//   - [WithPreInit]: make components able to Init (metadata, hooks)
//   - Init (via [AppCtx.Conf]): resources ready, no business traffic
//   - WithContext (via Conf inject): capabilities into ctx
//   - [WithPreRun]: process-level prep with a complete ctx
//   - [WithServe]: long-lived services (preferred for daemon apps)
//   - Main (constructor arg): process owner after Serve starts
//
// In [AppCtx.Conf], after writing defaults: [AppOption.PreInit] → Init → inject.
// On `run`: [AppOption.PreRun] → [AppOption.Serve] (async) → Main:
//
//   - [WithPreInit]: before Init in Conf (metadata, hooks, …)
//   - [WithPreRun]: ordered process prep (global config, inject context, …)
//   - [WithServe]: long-lived services (HTTP server, cron jobs, …)
//   - Main: decides wait vs exit (signal / Serves done / oneshot); must
//     [AppCtx.Close]. Framework does not infer Main from Serves emptiness.
//   - [WithClose]: extra close hooks; Conf components close automatically
//     in [AppCtx.Close] and do not need registration
//
// Prefer serve-first: put long-lived work in [WithServe]; use
// [AppCtx.AddCommand] for oneshot CLI tools. Main still owns when to exit.
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
