package types

import "time"

func Cost() func() time.Duration {
	ts := time.Now()
	return func() time.Duration {
		return time.Since(ts)
	}
}
