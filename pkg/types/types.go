package types

import "context"

type Runner interface {
	Run(ctx context.Context) error
}

type Defaulter interface {
	SetDefault()
}

type CanShutdown interface {
	Shutdown(ctx context.Context) error
}
