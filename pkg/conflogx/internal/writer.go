package internal

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const rollingTimeLayout = "20060102-15"

type fixedWriter struct {
	file *os.File
}

func openFixedWriter(dir, name string) (io.Writer, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filepath.Join(dir, name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	return &fixedWriter{file: file}, nil
}

func (w *fixedWriter) Write(p []byte) (int, error) {
	return w.file.Write(p)
}

type rollingWriter struct {
	dir  string
	now  func() time.Time
	mu   sync.Mutex
	file *os.File
	key  string
	idx  int
}

func openRollingWriter(dir string) (io.Writer, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	w := &rollingWriter{dir: dir, now: time.Now}
	if err := w.openCurrent(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *rollingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	key := rollingKey(w.now())
	if w.file == nil || w.key != key {
		if err := w.switchFile(key); err != nil {
			return 0, err
		}
	}
	return w.file.Write(p)
}

func (w *rollingWriter) switchFile(key string) error {
	if w.file != nil {
		_ = w.file.Close()
		w.file = nil
	}

	idx := latestRollingIndex(w.dir, key)
	file, err := os.OpenFile(rollingFilePath(w.dir, key, idx), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}

	w.key = key
	w.idx = idx
	w.file = file
	return nil
}

func (w *rollingWriter) openCurrent() error {
	return w.switchFile(rollingKey(w.now()))
}

func rollingKey(t time.Time) string {
	return t.Format(rollingTimeLayout)
}

func rollingFileName(key string, index int) string {
	return fmt.Sprintf("%s-%d.log", key, index)
}

func rollingFilePath(dir, key string, index int) string {
	return filepath.Join(dir, rollingFileName(key, index))
}

func latestRollingIndex(dir, key string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}

	prefix := key + "-"
	maxIdx := -1
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".log") {
			continue
		}
		raw := strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".log")
		idx, err := strconv.Atoi(raw)
		if err != nil || idx < 0 {
			continue
		}
		if idx > maxIdx {
			maxIdx = idx
		}
	}
	if maxIdx < 0 {
		return 0
	}
	return maxIdx
}
