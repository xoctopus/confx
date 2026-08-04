package appx

import "os"

// Runtime is the deployment environment label stored on [Meta].
type Runtime string

const (
	// RUNTIME_PROD is the production environment.
	RUNTIME_PROD Runtime = "PROD"
	// RUNTIME_STAGING is the staging environment.
	RUNTIME_STAGING Runtime = "STAGING"
	// RUNTIME_DEV is the local or development environment.
	RUNTIME_DEV Runtime = "DEV"
)

// String returns the runtime label.
func (r Runtime) String() string {
	return string(r)
}

// RuntimeEnvKey is the environment variable read by [GetRuntime].
const RuntimeEnvKey = "RUNTIME_ENV"

// GetRuntime returns the Runtime from [RuntimeEnvKey], or [RUNTIME_PROD] when
// unset or unrecognized.
func GetRuntime() Runtime {
	switch runtime := os.Getenv(RuntimeEnvKey); Runtime(runtime) {
	case RUNTIME_PROD, RUNTIME_STAGING, RUNTIME_DEV:
		return Runtime(runtime)
	default:
		return RUNTIME_PROD
	}
}
