package envconf

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/xoctopus/x/stringsx"
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
	return stringify(nil, '_', pw.paths...)
}

func (pw *PathWalker) CmdKey() string {
	return stringify(func(x string) string {
		x = strings.Replace(x, "_", "-", -1)
		return stringsx.LowerDashJoint(x)
	}, '-', pw.paths...)
}

func (pw *PathWalker) EnvKey() string {
	return stringify(func(x string) string {
		return strings.Replace(x, "-", "_", -1)
	}, '_', pw.paths...)
}

func stringify(h func(string) string, jointer rune, paths ...any) string {
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
			panic(errors.Errorf("unsupported type in path walker: %T", path))
		}
		if h != nil {
			s = h(s)
		}
		buf.WriteString(s)
	}

	return buf.String()
}
