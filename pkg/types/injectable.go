package types

import "context"

type Injectable interface {
	WithContext(ctx context.Context) context.Context
}
