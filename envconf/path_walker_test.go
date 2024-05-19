package envconf_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/sincospro/conf/envconf"
)

type stringer string

func (v stringer) String() string {
	return string(v)
}

func TestPathWalker(t *testing.T) {
	pw := envconf.NewPathWalker()

	pw.Enter("name")
	{
		NewWithT(t).Expect(pw.Paths()).To(Equal([]any{"name"}))
		NewWithT(t).Expect(pw.String()).To(Equal("name"))

		pw.Enter(1)
		{
			NewWithT(t).Expect(pw.Paths()).To(Equal([]any{"name", 1}))
			NewWithT(t).Expect(pw.String()).To(Equal("name_1"))

			pw.Enter("prop")
			{
				NewWithT(t).Expect(pw.Paths()).To(Equal([]any{"name", 1, "prop"}))
				NewWithT(t).Expect(pw.String()).To(Equal("name_1_prop"))
			}
			pw.Leave()
		}
		pw.Leave()

		pw.Enter(2)
		{
			NewWithT(t).Expect(pw.Paths()).To(Equal([]any{"name", 2}))
			NewWithT(t).Expect(pw.String()).To(Equal("name_2"))

			pw.Enter("prop")
			{
				NewWithT(t).Expect(pw.Paths()).To(Equal([]any{"name", 2, "prop"}))
				NewWithT(t).Expect(pw.String()).To(Equal("name_2_prop"))
			}
			pw.Leave()
		}
		pw.Leave()

		pw.Enter(stringer("3"))
		{
			NewWithT(t).Expect(pw.Paths()).To(Equal([]any{"name", stringer("3")}))
			NewWithT(t).Expect(pw.String()).To(Equal("name_3"))
		}
		pw.Leave()
	}
	pw.Leave()

	pw.Enter([]byte("unsupported type"))
	defer func() {
		t.Log(recover())
	}()
	_ = pw.String()
}
