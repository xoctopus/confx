package types_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/types"
)

func TestDurationParser(t *testing.T) {
	cases := map[string]time.Duration{
		// parse the short forms without spaces
		"1ns": 1 * time.Nanosecond,
		"1us": 1 * time.Microsecond,
		"1µs": 1 * time.Microsecond,
		"1ms": 1 * time.Millisecond,
		"1s":  1 * time.Second,
		"1m":  1 * time.Minute,
		"1h":  1 * time.Hour,
		"1hr": 1 * time.Hour,
		"1d":  24 * time.Hour,
		"1w":  7 * 24 * time.Hour,
		"1wk": 7 * 24 * time.Hour,

		// parse the long forms without spaces
		"1nanoseconds":  1 * time.Nanosecond,
		"1nanos":        1 * time.Nanosecond,
		"1microseconds": 1 * time.Microsecond,
		"1micros":       1 * time.Microsecond,
		"1millis":       1 * time.Millisecond,
		"1milliseconds": 1 * time.Millisecond,
		"1second":       1 * time.Second,
		"1sec":          1 * time.Second,
		"1min":          1 * time.Minute,
		"1minute":       1 * time.Minute,
		"1hour":         1 * time.Hour,
		"1day":          24 * time.Hour,
		"1week":         7 * 24 * time.Hour,

		// parse the short forms with spaces
		"1  ns": 1 * time.Nanosecond,
		"1  us": 1 * time.Microsecond,
		"1  µs": 1 * time.Microsecond,
		"1  ms": 1 * time.Millisecond,
		"1  s":  1 * time.Second,
		"1  m":  1 * time.Minute,
		"1  h":  1 * time.Hour,
		"1  hr": 1 * time.Hour,
		"1  d":  24 * time.Hour,
		"1  w":  7 * 24 * time.Hour,
		"1  wk": 7 * 24 * time.Hour,

		// parse the long forms without spaces
		"1  nanoseconds":  1 * time.Nanosecond,
		"1  nanos":        1 * time.Nanosecond,
		"1  microseconds": 1 * time.Microsecond,
		"1  micros":       1 * time.Microsecond,
		"1  millis":       1 * time.Millisecond,
		"1  milliseconds": 1 * time.Millisecond,
		"1  second":       1 * time.Second,
		"1  sec":          1 * time.Second,
		"1  min":          1 * time.Minute,
		"1  minute":       1 * time.Minute,
		"1  hour":         1 * time.Hour,
		"1  day":          24 * time.Hour,
		"1  week":         7 * 24 * time.Hour,
	}

	for str, dur := range cases {
		d, err := ParseDuration(str)
		Expect(t, err, Succeed())
		Expect(t, d, Equal(dur))
		x, err := time.ParseDuration(str)
		if err != nil {
			t.Log(str)
			continue
		}
		if x != d {
			t.Log(str)
			continue
		}
		Expect(t, x, Equal(d))
	}
}

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

func TestDuration(t *testing.T) {
	ds := []time.Duration{
		time.Second,
		10001,
		10001000,
		time.Hour,
		24 * time.Hour,
	}

	for _, d := range ds {
		dd := types.Duration(d)
		Expect(t, dd.IsZero(), BeFalse())
		data, err := dd.MarshalText()
		Expect(t, err, Succeed())
		Expect(t, string(data), Equal(dd.String()))

		dd2 := types.Duration(0)
		Expect(t, dd2.UnmarshalText([]byte(dd.String())), Succeed())
	}
}
