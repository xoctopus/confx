package internal

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/xoctopus/x/testx"
)

func TestRollingWriter(t *testing.T) {
	dir := t.TempDir()
	start := time.Date(2026, 8, 19, 14, 10, 0, 0, time.UTC)
	current := start

	w := &rollingWriter{
		dir: dir,
		now: func() time.Time { return current },
	}
	Expect(t, w.openCurrent(), Succeed())

	_, err := w.Write([]byte("hour-14"))
	Expect(t, err, Succeed())

	current = time.Date(2026, 8, 19, 15, 0, 0, 0, time.UTC)
	_, err = w.Write([]byte("hour-15"))
	Expect(t, err, Succeed())

	data14, err := os.ReadFile(filepath.Join(dir, "20260819-14-0.log"))
	Expect(t, err, Succeed())
	Expect(t, string(data14), Equal("hour-14"))

	data15, err := os.ReadFile(filepath.Join(dir, "20260819-15-0.log"))
	Expect(t, err, Succeed())
	Expect(t, string(data15), Equal("hour-15"))
}

func TestLatestRollingIndex(t *testing.T) {
	dir := t.TempDir()
	key := "20260819-15"

	for _, name := range []string{
		"20260819-14-0.log",
		"20260819-15-0.log",
		"20260819-15-2.log",
		"invalid.log",
	} {
		Expect(t, os.WriteFile(filepath.Join(dir, name), nil, 0o644), Succeed())
	}

	Expect(t, latestRollingIndex(dir, key), Equal(2))
	Expect(t, latestRollingIndex(dir, "20260819-99"), Equal(0))
}
