package confapp

import "os"

type Runtime string

const (
	RUNTIME_PROD    Runtime = "PROD"
	RUNTIME_STAGING Runtime = "STAGING"
	RUNTIME_DEV     Runtime = "DEV"
)

func (r Runtime) String() string {
	return string(r)
}

const RuntimeKey = "GO_RUNTIME"

func GetRuntime() Runtime {
	switch runtime := os.Getenv(RuntimeKey); Runtime(runtime) {
	case RUNTIME_PROD, RUNTIME_STAGING, RUNTIME_DEV:
		return Runtime(runtime)
	default:
		return RUNTIME_PROD
	}
}
