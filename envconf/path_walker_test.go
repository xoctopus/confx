package envconf_test

import (
	"testing"

	. "github.com/onsi/gomega"

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
		NewWithT(t).Expect(pw.Paths()).To(Equal([]any{"Group"}))
		NewWithT(t).Expect(pw.String()).To(Equal("Group"))

		pw.Enter(1)
		{
			NewWithT(t).Expect(pw.Paths()).To(Equal([]any{"Group", 1}))
			NewWithT(t).Expect(pw.String()).To(Equal("Group_1"))
			NewWithT(t).Expect(pw.CmdKey()).To(Equal("group-1"))
			NewWithT(t).Expect(pw.EnvKey()).To(Equal("Group_1"))

			pw.Enter("prop")
			{
				NewWithT(t).Expect(pw.Paths()).To(Equal([]any{"Group", 1, "prop"}))
				NewWithT(t).Expect(pw.String()).To(Equal("Group_1_prop"))
				NewWithT(t).Expect(pw.CmdKey()).To(Equal("group-1-prop"))
				NewWithT(t).Expect(pw.EnvKey()).To(Equal("Group_1_prop"))
			}
			pw.Leave()
		}
		pw.Leave()

		pw.Enter(stringer("Sub"))
		{
			NewWithT(t).Expect(pw.Paths()).To(Equal([]any{"Group", stringer("Sub")}))
			NewWithT(t).Expect(pw.String()).To(Equal("Group_Sub"))
			NewWithT(t).Expect(pw.EnvKey()).To(Equal("Group_Sub"))
			NewWithT(t).Expect(pw.CmdKey()).To(Equal("group-sub"))
		}
		pw.Leave()

		pw.Enter("UpperCamelSub")
		{
			NewWithT(t).Expect(pw.Paths()).To(Equal([]any{"Group", "UpperCamelSub"}))
			NewWithT(t).Expect(pw.String()).To(Equal("Group_UpperCamelSub"))
			NewWithT(t).Expect(pw.EnvKey()).To(Equal("Group_UpperCamelSub"))
			NewWithT(t).Expect(pw.CmdKey()).To(Equal("group-upper-camel-sub"))
		}
		pw.Leave()

		pw.Enter("lowerCamelSub")
		{
			NewWithT(t).Expect(pw.Paths()).To(Equal([]any{"Group", "lowerCamelSub"}))
			NewWithT(t).Expect(pw.String()).To(Equal("Group_lowerCamelSub"))
			NewWithT(t).Expect(pw.EnvKey()).To(Equal("Group_lowerCamelSub"))
			NewWithT(t).Expect(pw.CmdKey()).To(Equal("group-lower-camel-sub"))
		}
		pw.Leave()

		pw.Enter("Dash-Sub")
		{
			NewWithT(t).Expect(pw.Paths()).To(Equal([]any{"Group", "Dash-Sub"}))
			NewWithT(t).Expect(pw.String()).To(Equal("Group_Dash-Sub"))
			NewWithT(t).Expect(pw.EnvKey()).To(Equal("Group_Dash_Sub"))
			NewWithT(t).Expect(pw.CmdKey()).To(Equal("group-dash-sub"))
		}
		pw.Leave()

		pw.Enter("Underlined_Sub")
		{
			NewWithT(t).Expect(pw.Paths()).To(Equal([]any{"Group", "Underlined_Sub"}))
			NewWithT(t).Expect(pw.String()).To(Equal("Group_Underlined_Sub"))
			NewWithT(t).Expect(pw.EnvKey()).To(Equal("Group_Underlined_Sub"))
			NewWithT(t).Expect(pw.CmdKey()).To(Equal("group-underlined-sub"))
		}
		pw.Leave()
	}
	pw.Leave()

	pw.Enter([]byte("unsupported type"))
	defer func() {
		e := recover().(error).Error()
		NewWithT(t).Expect(e).To(HavePrefix("unsupported type in path walker:"))
	}()
	_ = pw.String()
}
