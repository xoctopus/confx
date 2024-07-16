package confcmd_test

import (
	"bytes"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/xoctopus/confx/confcmd"
)

func TestNewEnv(t *testing.T) {
	env := confcmd.NewEnv("prefix")

	t.Setenv("PREFIX__KEY", "100")
	NewWithT(t).Expect(env.Prefix()).To(Equal("prefix"))
	NewWithT(t).Expect(env.Var("key")).To(Equal("100"))

	env = confcmd.NewEnv("")
	t.Setenv("KEY", "101")
	NewWithT(t).Expect(env.Prefix()).To(Equal(""))
	NewWithT(t).Expect(env.Var("key")).To(Equal("101"))
}

type TestExecutor struct {
	Complex
	Embed

	*confcmd.Env
}

func (e *TestExecutor) Use() string { return "exec" }

func (e *TestExecutor) Short() string { return "a demo executor" }

func (e *TestExecutor) Long() string { return "a demo executor long description" }

func (e *TestExecutor) Example() string { return "a demo executor example" }

func (e *TestExecutor) Exec(cmd *cobra.Command, args ...string) error {
	cmd.Println("args")
	for _, arg := range args {
		cmd.Printf("\t%s\n", arg)
	}
	cmd.Println("flags")
	for _, name := range []string{
		"big-int",
		"has-default",
		"has-help-usage",
		"has-i18n-id",
		"integers",
		"no-opt-default",
		"persistent",
		"ss",
		"strings",
	} {
		f := cmd.Flag(name)
		if f == nil {
			continue
		}
		cmd.Printf("\t%s: %s\n", name, f.Value.String())
	}
	return nil
}

func ExampleNewCommand() {
	exec := &TestExecutor{
		Env: confcmd.NewEnv("test"),
	}

	buf := bytes.NewBuffer(nil)
	cmd := confcmd.NewCommand(exec, localizer)
	cmd.SetErr(buf)
	cmd.SetOut(buf)

	cmd.SetArgs([]string{
		"--big-int", "101",
		"--has-default", "102",
		"--has-help-usage", "103",
		"--has-i18n-id", "104",
		"--integers", "105 106 107",
		"--no-opt-default",
		"--persistent", "108.001",
		"--require", "109",
		"--ss", "110",
		"--strings", "111 112 113",
		"arg1",
		"arg2",
	})
	_ = cmd.Execute()
	fmt.Print(buf.String())

	// Output:
	// args
	// 	arg1
	// 	arg2
	// flags
	// 	big-int: 101
	// 	has-default: 102
	// 	has-help-usage: 103
	// 	has-i18n-id: 104
	// 	integers: [105 106 107]
	// 	no-opt-default: true
	// 	persistent: 108.001
	// 	ss: 110
	// 	strings: [111 112 113]
}
