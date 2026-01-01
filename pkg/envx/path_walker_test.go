package envx_test

import (
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/envx"
)

type stringer string

func (v stringer) String() string {
	return string(v)
}

func TestPathWalker(t *testing.T) {
	pw := envx.NewPathWalker()

	pw.Enter("Group")
	{
		Expect(t, pw.Paths(), Equal([]any{"Group"}))
		Expect(t, pw.String(), Equal("Group"))

		pw.Enter(1)
		{
			Expect(t, pw.Paths(), Equal([]any{"Group", 1}))
			Expect(t, pw.String(), Equal("Group_1"))

			pw.Enter("prop")
			{
				Expect(t, pw.Paths(), Equal([]any{"Group", 1, "prop"}))
				Expect(t, pw.String(), Equal("Group_1_prop"))
			}
			pw.Leave()
		}
		pw.Leave()

		pw.Enter(stringer("Sub"))
		{
			Expect(t, pw.Paths(), Equal([]any{"Group", stringer("Sub")}))
			Expect(t, pw.String(), Equal("Group_Sub"))
		}
		pw.Leave()

		pw.Enter("UpperCamelSub")
		{
			Expect(t, pw.Paths(), Equal([]any{"Group", "UpperCamelSub"}))
			Expect(t, pw.String(), Equal("Group_UpperCamelSub"))
		}
		pw.Leave()

		pw.Enter("lowerCamelSub")
		{
			Expect(t, pw.Paths(), Equal([]any{"Group", "lowerCamelSub"}))
			Expect(t, pw.String(), Equal("Group_lowerCamelSub"))
		}
		pw.Leave()

		pw.Enter("Dash-Sub")
		{
			Expect(t, pw.Paths(), Equal([]any{"Group", "Dash-Sub"}))
			Expect(t, pw.String(), Equal("Group_Dash-Sub"))
		}
		pw.Leave()

		pw.Enter("Underlined_Sub")
		{
			Expect(t, pw.Paths(), Equal([]any{"Group", "Underlined_Sub"}))
			Expect(t, pw.String(), Equal("Group_Underlined_Sub"))
		}
		pw.Leave()
	}
	pw.Leave()

	pw.Enter([]byte("unsupported type"))
	ExpectPanic[error](t, func() { _ = pw.String() }, ErrorContains("unsupported type in path walker:"))
}
