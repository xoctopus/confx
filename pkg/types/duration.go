package types

import (
	"time"
)

type Duration time.Duration

func (d Duration) IsZero() bool {
	return d == 0
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

func (d *Duration) UnmarshalText(data []byte) error { // validation is performed later on
	dd, err := time.ParseDuration(string(data))
	if err != nil {
		return err
	}
	*d = Duration(dd)
	return nil
}

func (d Duration) String() string {
	return time.Duration(d).String()
}
