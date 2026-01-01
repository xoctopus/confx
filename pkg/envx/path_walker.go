package envx

import (
	"bytes"
	"fmt"
	"strconv"
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
	return stringify('_', pw.paths...)
}

func stringify(jointer rune, paths ...any) string {
	buf := bytes.NewBuffer(nil)

	for idx, path := range paths {
		if idx > 0 {
			buf.WriteRune(jointer)
		}

		s := ""
		switch v := path.(type) {
		case string:
			s = v
		case int:
			s = strconv.Itoa(v)
		case fmt.Stringer:
			s = v.String()
		default:
			panic(fmt.Errorf("unsupported type in path walker: %T", path))
		}
		buf.WriteString(s)
	}

	return buf.String()
}
