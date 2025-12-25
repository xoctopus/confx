package envconf_test

import (
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/envconf"
)

type stringer string

func (v stringer) String() string {
	return string(v)
}

func TestPathWalker(t *testing.T) {
	pw := envconf.NewPathWalker()

	pw.Enter("Group")
	{
		Expect(t, pw.Paths(), Equal([]any{"Group"}))
		Expect(t, pw.String(), Equal("Group"))

		pw.Enter(1)
		{
			Expect(t, pw.Paths(), Equal([]any{"Group", 1}))
			Expect(t, pw.String(), Equal("Group_1"))
			Expect(t, pw.CmdKey(), Equal("group-1"))
			Expect(t, pw.EnvKey(), Equal("Group_1"))

			pw.Enter("prop")
			{
				Expect(t, pw.Paths(), Equal([]any{"Group", 1, "prop"}))
				Expect(t, pw.String(), Equal("Group_1_prop"))
				Expect(t, pw.CmdKey(), Equal("group-1-prop"))
				Expect(t, pw.EnvKey(), Equal("Group_1_prop"))
			}
			pw.Leave()
		}
		pw.Leave()

		pw.Enter(stringer("Sub"))
		{
			Expect(t, pw.Paths(), Equal([]any{"Group", stringer("Sub")}))
			Expect(t, pw.String(), Equal("Group_Sub"))
			Expect(t, pw.EnvKey(), Equal("Group_Sub"))
			Expect(t, pw.CmdKey(), Equal("group-sub"))
		}
		pw.Leave()

		pw.Enter("UpperCamelSub")
		{
			Expect(t, pw.Paths(), Equal([]any{"Group", "UpperCamelSub"}))
			Expect(t, pw.String(), Equal("Group_UpperCamelSub"))
			Expect(t, pw.EnvKey(), Equal("Group_UpperCamelSub"))
			Expect(t, pw.CmdKey(), Equal("group-upper-camel-sub"))
		}
		pw.Leave()

		pw.Enter("lowerCamelSub")
		{
			Expect(t, pw.Paths(), Equal([]any{"Group", "lowerCamelSub"}))
			Expect(t, pw.String(), Equal("Group_lowerCamelSub"))
			Expect(t, pw.EnvKey(), Equal("Group_lowerCamelSub"))
			Expect(t, pw.CmdKey(), Equal("group-lower-camel-sub"))
		}
		pw.Leave()

		pw.Enter("Dash-Sub")
		{
			Expect(t, pw.Paths(), Equal([]any{"Group", "Dash-Sub"}))
			Expect(t, pw.String(), Equal("Group_Dash-Sub"))
			Expect(t, pw.EnvKey(), Equal("Group_Dash_Sub"))
			Expect(t, pw.CmdKey(), Equal("group-dash-sub"))
		}
		pw.Leave()

		pw.Enter("Underlined_Sub")
		{
			Expect(t, pw.Paths(), Equal([]any{"Group", "Underlined_Sub"}))
			Expect(t, pw.String(), Equal("Group_Underlined_Sub"))
			Expect(t, pw.EnvKey(), Equal("Group_Underlined_Sub"))
			Expect(t, pw.CmdKey(), Equal("group-underlined-sub"))
		}
		pw.Leave()
	}
	pw.Leave()

	pw.Enter([]byte("unsupported type"))
	ExpectPanic[error](t, func() { _ = pw.String() }, ErrorContains("unsupported type in path walker:"))
}
