package types

import "context"

type Servable interface {
	Serve(ctx context.Context) error

	Shutdownable
}
