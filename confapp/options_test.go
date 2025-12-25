package confapp_test

import (
	"testing"
	"time"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/confapp"
)

func TestMeta(t *testing.T) {
	m1 := confapp.DefaultMeta
	m2 := confapp.Meta{
		Name:     "test",
		Feature:  "test/abc",
		Version:  "v1.1.1",
		CommitID: "abcdef0",
		Date:     time.Now().Format("200601021504"),
		Runtime:  "DEV",
	}

	m1.Overwrite(m2)
	Expect(t, m1.String(), Equal(m2.String()))
}

func TestAppOption(t *testing.T) {
	opt := confapp.AppOption{
		Meta: confapp.DefaultMeta,
	}
	Expect(t, opt.NeedAttach(), BeFalse())
	opt.AppendPreRunners(func() {})
	opt.AppendBatchRunners(func() {})
	opt.PreRun()
}
