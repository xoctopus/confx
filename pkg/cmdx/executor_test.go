package cmdx_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/cmdx"
)

type Basics struct {
	Required   string   `cmd:",require"`
	Persistent *float32 `cmd:",persist"`
	HasDefault string   `cmd:",default='has-default'"`
	Shorthand  int      `cmd:"short-hand,short=s"`
	Skipped    any      `cmd:"-"`
	unexported any
}

func (b *Basics) DocOf(names ...string) ([]string, bool) {
	if len(names) == 0 {
		return []string{"basic datatype flags"}, true
	}
	switch names[0] {
	case "Required":
		return []string{"string", "case [required]"}, true
	case "Persistent":
		return []string{"*float32", "case [persistent]"}, true
	case "HasDefault":
		return []string{"string", "case [has-no-option-default]"}, true
	case "Shorthand":
		return []string{"int", "case [has-shorthand to 's']"}, true
	case "Skipped":
		return []string{"any", "case [skip flag marked as '-']"}, true
	}
	return []string{}, false
}

type Composited struct {
	StringArray  []*string
	IntegerArray []int64
	BigInt       *big.Int // big.Int implements text.Arshaler
}

func (c *Composited) DocOf(names ...string) ([]string, bool) {
	if len(names) == 0 {
		return []string{"composited datatype flags"}, true
	}
	switch names[0] {
	case "StringArray":
		return []string{"[]*string", "case [strings]"}, true
	case "IntegerArray":
		return []string{"[]int64", "case [integers]"}, true
	case "BigInt":
		return []string{"*big.Int", "case [TextMarshaler]"}, true
	}
	return []string{}, false
}

type Prefixed struct {
	Required    string   `cmd:",require"`
	Persistent  *float32 `cmd:",persist"`
	HasDefault  string   `cmd:",default='100 101 102'"`
	HasNoOptDef []string `cmd:",noopdef='103 104 105'"`
}

func (c *Prefixed) DocOf(names ...string) ([]string, bool) {
	if len(names) == 0 {
		return []string{"prefixed flags"}, true
	}
	switch names[0] {
	case "Required":
		return []string{"string", "case [prefixed,require]"}, true
	case "Persistent":
		return []string{"*float32", "case [prefixed,persistent]"}, true
	case "HasDefault":
		return []string{"string", "case [prefixed,default]"}, true
	case "HasNoOptDef":
		return []string{"[]string", "case [prefixed,no option default]"}, true
	}
	return []string{}, false
}

type Inline struct {
	Required   string   `json:"required"`
	Persistent *float32 `json:"persistent"`
	HasDefault string   `json:"hasDefault"`
}

func (v *Inline) UnmarshalText(text []byte) error {
	tmp := struct {
		Required   string   `json:"required"`
		Persistent *float32 `json:"persistent"`
		HasDefault string   `json:"hasDefault"`
	}{}
	if err := json.Unmarshal(text, &tmp); err != nil {
		return err
	}
	v.Required = tmp.Required
	v.HasDefault = tmp.HasDefault
	v.Persistent = tmp.Persistent
	return nil
}

func (v Inline) MarshalText() ([]byte, error) {
	return json.Marshal(struct {
		Required   string   `json:"required"`
		Persistent *float32 `json:"persistent"`
		HasDefault string   `json:"hasDefault"`
	}{
		Required:   v.Required,
		Persistent: v.Persistent,
		HasDefault: v.HasDefault,
	})
}

type TestExecutor struct {
	Basics
	*Composited
	Prefixed `cmd:"prefixed"`
	Inline
	Global string
}

func (t *TestExecutor) DocOf(names ...string) ([]string, bool) {
	if len(names) == 0 {
		return []string{"Test Executor"}, true
	}
	switch names[0] {
	case "Global":
		return []string{"global string"}, true
	case "Inline":
		return []string{"TextArshaler"}, true
	}
	if names[0] == "Global" {
	}

	return []string{}, false
}

func (*TestExecutor) Exec(cmd *cobra.Command, args ...string) error {
	cmd.Println("args:")
	for _, arg := range args {
		cmd.Printf("    %s\n", arg)
	}
	cmd.Println()

	cmd.Println("flags")

	names := []string{
		// TestBasics
		"required",
		"persistent",
		"has-default",
		"short-hand",
		// *TestComposited
		"string-array",
		"integer-array",
		"big-int",
		// TestPrefixed
		"prefixed-required",
		"prefixed-persistent",
		"prefixed-has-default",
		"prefixed-has-no-opt-def",
		// TestInline
		"inline",
		// Global
		"global",
		// not registered
		"must-not-found",
	}
	for _, name := range names {
		f := cmd.Flag(name)

		if f == nil {
			cmd.Printf("    %-24s flag is not registered\n", name+":")
			continue
		}
		v := f.Value.String()
		if v == "" {
			v = f.NoOptDefVal
		}
		cmd.Printf("    %-24s %s\n", name+":", v)
	}

	return nil
}

func NewCommand(v cmdx.Executor, buf io.ReadWriter, args ...string) cmdx.Command {
	c := cmdx.NewCommand("demo", v)

	c.WithShort("this is a short description for executor `demo`")
	c.WithLong("this is a long description for executor `demo`")
	c.WithExample("this is an example for executor `demo`")

	cmd := c.Cmd()

	cmd.SetErr(buf)
	cmd.SetOut(buf)
	cmd.SetArgs(args)
	return c
}

func ExampleExecutor() {
	_ = os.Setenv("DEMO__REQUIRED", "true")

	buf := bytes.NewBuffer(nil)
	cmd := NewCommand(
		&TestExecutor{},
		buf,
		// TestBasics
		"--required", "required",
		"--persistent", "100.001",
		// "--has-default", "",
		"-s", "101",
		// *TestComposited
		"--string-array", "str1 str2 str3",
		"--integer-array", "100 101 102",
		"--big-int", "1111111111111111111111111111111111",
		// TestPrefixed
		"--prefixed-required", "prefixed required",
		"--prefixed-persistent", "100.02",
		"--prefixed-has-default", "overwrite default",
		"--prefixed-has-no-opt-def",
		// TestInline
		"--inline", `{"required": "required", "persistent": 100.002, "hasDefault": "default string"}`,
		// Global
		"--global", "global string",
		// append arguments
		"arg1",
		"arg2",
	)

	cmd.Cmd().Printf("command: %s\n\n", cmd.Use())
	_ = cmd.Cmd().Execute()
	fmt.Print(buf.String())

	// Output:
	// command: demo
	//
	// args:
	//     arg1
	//     arg2
	//
	// flags
	//     required:                required
	//     persistent:              100.001
	//     has-default:             has-default
	//     short-hand:              101
	//     string-array:            [str1 str2 str3]
	//     integer-array:           [100 101 102]
	//     big-int:                 1111111111111111111111111111111111
	//     prefixed-required:       prefixed required
	//     prefixed-persistent:     100.02
	//     prefixed-has-default:    overwrite default
	//     prefixed-has-no-opt-def: [103 104 105]
	//     inline:                  {"required":"required","persistent":100.002,"hasDefault":"default string"}
	//     global:                  global string
	//     must-not-found:          flag is not registered
}

func ExampleExecutor_help() {
	buf := bytes.NewBuffer(nil)
	cmd := NewCommand(&TestExecutor{}, buf, "--help")

	_ = cmd.Cmd().Execute()
	log.Print("\n" + buf.String())

	//Output:
}

type MockExecutor struct {
	Integers []int
	Inline
}

func (*MockExecutor) Exec(cmd *cobra.Command, args ...string) error { return nil }

func TestFlag_Set(t *testing.T) {
	t.Run("FailedUnmarshalText", func(t *testing.T) {
		t.Setenv("DEMO__INLINE", "[]")
		defer os.Unsetenv("DEMO__INLINE")

		buf := bytes.NewBuffer(nil)
		testx.ExpectPanic[error](t, func() {
			_ = NewCommand(&MockExecutor{}, buf)
		})
	})
	t.Run("FailedUnmarshalFields", func(t *testing.T) {
		t.Setenv("DEMO__INTEGERS", "1 2 abc")
		defer os.Unsetenv("DEMO__INTEGERS")

		buf := bytes.NewBuffer(nil)
		testx.ExpectPanic[error](t, func() {
			_ = NewCommand(&MockExecutor{}, buf)
		})
	})
}
