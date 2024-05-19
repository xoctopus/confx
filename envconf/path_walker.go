package envconf

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

func NewPathWalker() *PathWalker {
	return &PathWalker{}
}

type PathWalker struct {
	paths []any
}

func (pw *PathWalker) Enter(i any) {
	pw.paths = append(pw.paths, i)
}

func (pw *PathWalker) Leave() {
	pw.paths = pw.paths[:len(pw.paths)-1]
}

func (pw *PathWalker) Paths() []any {
	return pw.paths
}

func (pw *PathWalker) String() string {
	return StringifyPath(pw.paths...)
}

func StringifyPath(paths ...any) string {
	buf := bytes.NewBuffer(nil)

	for idx, path := range paths {
		if idx > 0 {
			buf.WriteRune('_')
		}

		switch v := path.(type) {
		case string:
			buf.WriteString(v)
		case int:
			buf.WriteString(strconv.Itoa(v))
		case fmt.Stringer:
			buf.WriteString(v.String())
		default:
			panic(errors.Errorf("unsupported type in path walker: %T", path))
		}
	}

	return buf.String()
}
