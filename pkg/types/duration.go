package types

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	units = [][]string{
		{"ns", "nano"},
		{"us", "µs", "micro"},
		{"ms", "milli"},
		{"s", "sec"},
		{"m", "min"},
		{"h", "hr", "hour"},
		{"d", "day"},
		{"w", "wk", "week"},
	}

	dus = map[string]time.Duration{
		"ns": time.Nanosecond,
		"us": time.Microsecond,
		"ms": time.Millisecond,
		"s":  time.Second,
		"m":  time.Minute,
		"h":  time.Hour,
		"d":  24 * time.Hour,
		"w":  7 * 24 * time.Hour,
	}

	regexDuration = regexp.MustCompile(`((\d+)\s*([A-Za-zµ]+))`)
)

func ParseDuration(s string) (time.Duration, error) {
	if du, err := time.ParseDuration(s); err == nil {
		return du, nil
	}

	var du time.Duration
	ok := false
	for _, match := range regexDuration.FindAllStringSubmatch(s, -1) {
		factor, err := strconv.Atoi(match[2]) // converts string to int
		if err != nil {
			return 0, err
		}
		unit := strings.ToLower(strings.TrimSpace(match[3]))

		for _, variants := range units {
			last := len(variants) - 1
			u := dus[variants[0]]

			for i, variant := range variants {
				if (last == i && strings.HasPrefix(unit, variant)) ||
					strings.EqualFold(variant, unit) {
					ok = true
					du += time.Duration(factor) * u
				}
			}
		}
	}

	if ok {
		return du, nil
	}
	return 0, fmt.Errorf("failed to parse %s as duration", s)
}

func IsDuration(str string) bool {
	_, err := ParseDuration(str)
	return err == nil
}

type Duration time.Duration

func (d Duration) IsZero() bool {
	return d == 0
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

func (d *Duration) UnmarshalText(data []byte) error { // validation is performed later on
	dd, err := ParseDuration(string(data))
	if err != nil {
		return err
	}
	*d = Duration(dd)
	return nil
}

// func (d *Duration) Scan(raw interface{}) error {
// 	switch v := raw.(type) {
// 	case int64:
// 		*d = Duration(v)
// 	case float64:
// 		*d = Duration(int64(v))
// 	case nil:
// 		*d = Duration(0)
// 	default:
// 		return fmt.Errorf("cannot sql.Scan() strfmt.Duration from: %#v", v)
// 	}
//
// 	return nil
// }

func (d Duration) String() string {
	return time.Duration(d).String()
}
