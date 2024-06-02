package confcmd_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	. "github.com/xoctopus/confx/confcmd"
)

type TestStruct struct {
	unexported           int
	Ignored              string   `cmd:"-"`
	IsRequired           []string `cmd:",required"`
	Debug                bool     `cmd:",p,nop=1"` // has no option default and as persistent flag
	OverwrittenFlagName  bool     `cmd:"is-bool"`
	OverwrittenEnvKey    string   `env:"is-string"`
	DisableEnvInject     int32    `env:"-"`
	HasMultiLangHelp     []int    `help:"help default" help.en:"help en" help.zh:"中文帮助"`
	HasDefaultValue      []float64
	OverwrittenGroupName ***Basics `cmd:"values"`
	SlicesGroup          **Slices  // flags group
	Basics                         // inline anonymous
	Slices               `cmd:"-"` // skip

	// Executor enhancements
	*FlagSet         `cmd:"-"`
	*MultiLangHelper `cmd:"-"`
	*EnvInjector     `cmd:"-"`
}

func (e *TestStruct) Use() string { return "cmd use" }

func (e *TestStruct) Short() string { return "cmd short" }

func (e *TestStruct) Exec(cmd *cobra.Command, args ...string) error {
	cmd.Printf("is-required: %s\n", cmd.Flag("is-required").Value.String())
	cmd.Printf("string:      %s\n", cmd.Flag("string").Value.String())
	return nil
}

func (e *TestStruct) Long() string { return "cmd long description" }

func (e *TestStruct) Example() string { return "cmd usage example" }

type Basics struct {
	IntPtr  *int
	Int     int
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	UintPtr *uint
	Uint    uint
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Uint64  uint64
	Float32 float32
	Float64 float64
	String  string
	Boolean bool
}

type Slices struct {
	IntegerSlice     []int
	Integer32Slice   []int32
	Integer64Slice   []int64
	UnsignedIntegers []uint
	Float32Slice     []float32
	Float64Slice     []float64
	StringSlice      []string
	BooleanSlice     []bool
}

func TestParseFlags(t *testing.T) {
	t.Run("InvalidValue", func(t *testing.T) {
		defer func() {
			v := recover().(error)
			NewWithT(t).Expect(v).NotTo(BeNil())
			NewWithT(t).Expect(v.Error()).To(ContainSubstring("invalid input value"))
		}()
		ParseFlags(nil)
	})

	t.Run("NotStruct", func(t *testing.T) {
		defer func() {
			v := recover().(error)
			NewWithT(t).Expect(v).NotTo(BeNil())
			NewWithT(t).Expect(v.Error()).To(ContainSubstring("expect a struct value"))
		}()
		v := new(int)
		ParseFlags(v)
	})

	t.Run("CannotSet", func(t *testing.T) {
		defer func() {
			v := recover().(error)
			NewWithT(t).Expect(v).NotTo(BeNil())
			NewWithT(t).Expect(v.Error()).To(ContainSubstring("expect value can set"))
		}()
		v := struct{}{}
		ParseFlags(v)
	})

	t.Run("ParseStructFields", func(t *testing.T) {
		v := &TestStruct{
			HasDefaultValue: []float64{1, 2, 3},
		}

		flags := ParseFlags(v)

		find := func(field string) *Flag {
			for _, f := range flags {
				if f.Field() == field {
					return f
				}
			}
			return nil
		}

		NewWithT(t).Expect(find("unexported")).To(BeNil())
		NewWithT(t).Expect(find("Ignored")).To(BeNil())
		NewWithT(t).Expect(find("IsRequired").IsRequired()).To(BeTrue())
		NewWithT(t).Expect(find("OverwrittenFlagName").Name()).To(Equal("is-bool"))
		NewWithT(t).Expect(find("OverwrittenEnvKey").EnvKey("any")).To(Equal("ANY__IS_STRING"))
		NewWithT(t).Expect(find("OverwrittenEnvKey").EnvKey("")).To(Equal("IS_STRING"))
		NewWithT(t).Expect(find("DisableEnvInject").EnvKey("any")).To(Equal(""))
		NewWithT(t).Expect(find("HasMultiLangHelp").Help(FlagHelp)).To(Equal("help default"))
		NewWithT(t).Expect(find("HasMultiLangHelp").Help(LangEN)).To(Equal("help en"))
		NewWithT(t).Expect(find("HasMultiLangHelp").Help(LangZH)).To(Equal("中文帮助"))
		NewWithT(t).Expect(find("HasDefaultValue").DefaultValue()).To(Equal(v.HasDefaultValue))
		NewWithT(t).Expect(find("values.UintPtr")).NotTo(BeNil())
		NewWithT(t).Expect(find("SlicesGroup.IntegerSlice")).NotTo(BeNil())
		NewWithT(t).Expect(find("IntPtr")).NotTo(BeNil())
		NewWithT(t).Expect(find("Slices")).To(BeNil())
		NewWithT(t).Expect(find("Slices.UnsignedIntegers")).To(BeNil())
	})
}

func TestNewCommand(t *testing.T) {
	t.Run("FailedToAddFlag", func(t *testing.T) {
		t.Run("FlagNameConflict", func(t *testing.T) {
			defer func() {
				v := recover().(error)
				NewWithT(t).Expect(v.Error()).To(ContainSubstring("flag name conflict"))
			}()

			var v Executor = &struct {
				*TestStruct
				Int int // this flag name will conflict with TestStruct.Basics.Int
			}{
				TestStruct: &TestStruct{
					MultiLangHelper: NewDefaultMultiLangHelper(),
					FlagSet:         NewFlagSet(),
					EnvInjector:     NewEnvInjector("demo"),
				},
			}
			_ = NewCommand(v)
		})

		t.Run("EnvKeyConflict", func(t *testing.T) {
			defer func() {
				v := recover().(error)
				NewWithT(t).Expect(v.Error()).To(ContainSubstring("env key conflict"))
			}()

			var v Executor = &struct {
				*TestStruct
				UserDefine string `env:"is-string"` // this env key will conflict with TestStruct.OverwrittenEnvKey
			}{
				TestStruct: &TestStruct{
					MultiLangHelper: NewDefaultMultiLangHelper(),
					FlagSet:         NewFlagSet(),
					EnvInjector:     NewEnvInjector("demo"),
				},
			}
			_ = NewCommand(v)
		})

		t.Run("UnsupportedFlagType", func(t *testing.T) {
			defer func() {
				v := recover().(error)
				NewWithT(t).Expect(v.Error()).To(ContainSubstring("unsupported type"))
			}()

			var v Executor = &struct {
				*TestStruct
				Field []struct{ Int int }
			}{
				TestStruct: &TestStruct{
					MultiLangHelper: NewDefaultMultiLangHelper(),
					FlagSet:         NewFlagSet(),
					EnvInjector:     NewEnvInjector("demo"),
				},
				Field: []struct{ Int int }{{1}},
			}
			_ = NewCommand(v)
		})
	})
}

func ExampleNewCommand_from_args() {
	var v Executor = &TestStruct{
		HasDefaultValue: []float64{1, 2, 3},
		MultiLangHelper: NewDefaultMultiLangHelper(),
		FlagSet:         NewFlagSet(),
		EnvInjector:     NewEnvInjector("test"),
	}
	cmd := NewCommand(v)

	if v.Flag("is-required") == nil {
		return
	}
	flags := v.Flags()
	if flags["is-required"] == nil {
		return
	}

	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	{
		buf.Reset()
		cmd.SetArgs([]string{"--is-required", "a", "--is-required", "b", "--is-required", "c"})
		if err := cmd.Execute(); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(buf.String())
	}

	// Output:
	// is-required: [a b c]
	// string:
}

func ExampleNewCommand_from_env() {
	var v Executor = &TestStruct{
		HasDefaultValue: []float64{1, 2, 3},
		MultiLangHelper: NewDefaultMultiLangHelper(),
		FlagSet:         NewFlagSet(),
		EnvInjector:     NewEnvInjector("test"),
	}
	_ = os.Setenv("TEST__STRING", "i love you")

	cmd := NewCommand(v)

	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	{
		buf.Reset()
		cmd.SetArgs([]string{"--is-required", "d e f"})
		if err := cmd.Execute(); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(buf.String())
	}

	// Output:
	// is-required: [d e f]
	// string:      i love you
}

func ExampleNewCommand_output_flags() {
	var v Executor = &TestStruct{
		HasDefaultValue: []float64{1, 2, 3},
		MultiLangHelper: NewDefaultMultiLangHelper(),
		FlagSet:         NewFlagSet(),
		EnvInjector:     NewEnvInjector("demo"),
	}
	_ = NewCommand(v)

	injector := v.(CanInjectFromEnv)
	injector.SetPrefix("demo")
	v.SetHelpLang(LangZH)

	v.OutputFlagsHelp(v.HelpLang(), injector.Prefix())

	// Output:
	// ┌────────────────────────────────┬──────────┬───────────────────┬───────────────┬──────────────────────────────────────┬──────────────┐
	// │ flag name                      │ required │ no option default │ default value │ environment key                      │ help info    │
	// ├────────────────────────────────┼──────────┼───────────────────┼───────────────┼──────────────────────────────────────┼──────────────┤
	// │ boolean                        │          │ -                 │ false         │ DEMO__BOOLEAN                        │              │
	// │ debug                          │          │ 1                 │ false         │ DEMO__DEBUG                          │              │
	// │ disable-env-inject             │          │ -                 │ 0             │                                      │              │
	// │ float-32                       │          │ -                 │ 0             │ DEMO__FLOAT_32                       │              │
	// │ float-64                       │          │ -                 │ 0             │ DEMO__FLOAT_64                       │              │
	// │ has-default-value              │          │ -                 │ [1 2 3]       │ DEMO__HAS_DEFAULT_VALUE              │              │
	// │ has-multi-lang-help            │          │ -                 │ []            │ DEMO__HAS_MULTI_LANG_HELP            │ 中文帮助     │
	// │ int                            │          │ -                 │ 0             │ DEMO__INT                            │              │
	// │ int-16                         │          │ -                 │ 0             │ DEMO__INT_16                         │              │
	// │ int-32                         │          │ -                 │ 0             │ DEMO__INT_32                         │              │
	// │ int-64                         │          │ -                 │ 0             │ DEMO__INT_64                         │              │
	// │ int-8                          │          │ -                 │ 0             │ DEMO__INT_8                          │              │
	// │ int-ptr                        │          │ -                 │ 0             │ DEMO__INT_PTR                        │              │
	// │ is-bool                        │          │ -                 │ false         │ DEMO__OVERWRITTEN_FLAG_NAME          │              │
	// │ is-required                    │ yes      │ -                 │ []            │ DEMO__IS_REQUIRED                    │              │
	// │ overwritten-env-key            │          │ -                 │               │ DEMO__IS_STRING                      │              │
	// │ slices-group-boolean-slice     │          │ -                 │ []            │ DEMO__SLICES_GROUP_BOOLEAN_SLICE     │              │
	// │ slices-group-float-32-slice    │          │ -                 │ []            │ DEMO__SLICES_GROUP_FLOAT_32_SLICE    │              │
	// │ slices-group-float-64-slice    │          │ -                 │ []            │ DEMO__SLICES_GROUP_FLOAT_64_SLICE    │              │
	// │ slices-group-integer-32-slice  │          │ -                 │ []            │ DEMO__SLICES_GROUP_INTEGER_32_SLICE  │              │
	// │ slices-group-integer-64-slice  │          │ -                 │ []            │ DEMO__SLICES_GROUP_INTEGER_64_SLICE  │              │
	// │ slices-group-integer-slice     │          │ -                 │ []            │ DEMO__SLICES_GROUP_INTEGER_SLICE     │              │
	// │ slices-group-string-slice      │          │ -                 │ []            │ DEMO__SLICES_GROUP_STRING_SLICE      │              │
	// │ slices-group-unsigned-integers │          │ -                 │ []            │ DEMO__SLICES_GROUP_UNSIGNED_INTEGERS │              │
	// │ string                         │          │ -                 │               │ DEMO__STRING                         │              │
	// │ uint                           │          │ -                 │ 0             │ DEMO__UINT                           │              │
	// │ uint-16                        │          │ -                 │ 0             │ DEMO__UINT_16                        │              │
	// │ uint-32                        │          │ -                 │ 0             │ DEMO__UINT_32                        │              │
	// │ uint-64                        │          │ -                 │ 0             │ DEMO__UINT_64                        │              │
	// │ uint-8                         │          │ -                 │ 0             │ DEMO__UINT_8                         │              │
	// │ uint-ptr                       │          │ -                 │ 0             │ DEMO__UINT_PTR                       │              │
	// │ values-boolean                 │          │ -                 │ false         │ DEMO__VALUES_BOOLEAN                 │              │
	// │ values-float-32                │          │ -                 │ 0             │ DEMO__VALUES_FLOAT_32                │              │
	// │ values-float-64                │          │ -                 │ 0             │ DEMO__VALUES_FLOAT_64                │              │
	// │ values-int                     │          │ -                 │ 0             │ DEMO__VALUES_INT                     │              │
	// │ values-int-16                  │          │ -                 │ 0             │ DEMO__VALUES_INT_16                  │              │
	// │ values-int-32                  │          │ -                 │ 0             │ DEMO__VALUES_INT_32                  │              │
	// │ values-int-64                  │          │ -                 │ 0             │ DEMO__VALUES_INT_64                  │              │
	// │ values-int-8                   │          │ -                 │ 0             │ DEMO__VALUES_INT_8                   │              │
	// │ values-int-ptr                 │          │ -                 │ 0             │ DEMO__VALUES_INT_PTR                 │              │
	// │ values-string                  │          │ -                 │               │ DEMO__VALUES_STRING                  │              │
	// │ values-uint                    │          │ -                 │ 0             │ DEMO__VALUES_UINT                    │              │
	// │ values-uint-16                 │          │ -                 │ 0             │ DEMO__VALUES_UINT_16                 │              │
	// │ values-uint-32                 │          │ -                 │ 0             │ DEMO__VALUES_UINT_32                 │              │
	// │ values-uint-64                 │          │ -                 │ 0             │ DEMO__VALUES_UINT_64                 │              │
	// │ values-uint-8                  │          │ -                 │ 0             │ DEMO__VALUES_UINT_8                  │              │
	// │ values-uint-ptr                │          │ -                 │ 0             │ DEMO__VALUES_UINT_PTR                │              │
	// └────────────────────────────────┴──────────┴───────────────────┴───────────────┴──────────────────────────────────────┴──────────────┘
}
