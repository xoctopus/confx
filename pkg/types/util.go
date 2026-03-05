package types

import "time"

func Span() func() time.Duration {
	ts := time.Now()
	return func() time.Duration {
		return time.Since(ts)
	}
}
