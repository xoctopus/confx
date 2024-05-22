package confcmd_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sincospro/x/ptrx"
	"github.com/spf13/cobra"

	cmdx "github.com/sincospro/conf/confcmd"
)

var mustParsedFlags = []string{
	"multi-lang",
	"overwritten",
	"embed-integers",
	"embed-floats",
	"embed-strings",
	"int-ptr",
	"int",
	"int-8",
	"int-16",
	"int-32",
	"int-64",
	"uint-ptr",
	"uint",
	"uint-8",
	"uint-16",
	"uint-32",
	"uint-64",
	"float-32",
	"float-64",
	"string",
	"boolean",
}

var mustNotParsedFlags = []string{
	"skipped-unexported",
	"skipped-ignore",
	"be-overwritten",
}

type DemoOptions struct {
	MultiLang         int     `help.en:"multi language" help.zh:"多语言"`
	skippedUnexported any     `help.en:"any"  help.zh:"随便"`
	SkippedIgnored    any     `name:"-"`
	BeOverwritten     bool    `name:"overwritten"`
	IsRequired        string  `name:"is-required,required" help:"a" help.en:"b" help.zh:"帮助"`
	Embed             *Slices // embed struct
	Basics                    // anonymous struct
}

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
	Integers         []int64
	Floats           []float64
	Strings          []string
	UnsignedIntegers []uint64 // cobra unsupported
}

func ExampleNewCommand() {
	type Inline struct {
		InlineField int
	}
	var opt = &struct {
		DemoOptions
		HasDefaults string `name:"rewrite-tag"`
		IsRequired  bool   `name:",required"`
		HasHelp     string `help:"default describe" help.en:"en describe" help.zh:"中文帮助"`
		Embed       struct {
			Field int
		}
		Inline
	}{
		HasDefaults: "has-defaults",
	}

	cmd, err := cmdx.NewCommand(cmdx.ZH, opt)
	if err != nil {
		return
	}

	for _, _ = range []string{"rewrite-tag", "is-required", "has-help", "embed-field", "inline-field"} {
		// flag := cmd.Flags().Lookup(name)
		// if flag != nil {
		// 	_name := flag.Name
		// 	if flag.Deprecated
		// 	fmt.Printf("--%s: %s\n", flag.Name, flag.Usage, flag.DefValue)
		// 	if flag.Usage != "" {
		// 		(default: %s)
		// }
		// }
	}
	_ = cmd.Help()

	//demo short
	//
	//Usage:
	//  demo [flags]
	//
	//Flags:
	//      --boolean
	//      --embed-field int
	//      --embed-floats float64Slice    (default [])
	//      --embed-integers int64Slice    (default [])
	//      --embed-strings strings
	//      --float-32 float32
	//      --float-64 float
	//      --has-help string             中文帮助
	//      --inline-field int
	//      --int int
	//      --int-16 int16
	//      --int-32 int32
	//      --int-64 int
	//      --int-8 int8
	//      --int-ptr int
	//      --is-required
	//      --multi-lang int              多语言
	//      --overwritten
	//      --rewrite-tag string           (default "has-defaults")
	//      --string string
	//      --uint uint
	//      --uint-16 uint16
	//      --uint-32 uint32
	//      --uint-64 uint
	//      --uint-8 uint8
	//      --uint-ptr uint
}

func TestNewCommand(t *testing.T) {
	cmd, err := cmdx.NewCommand(cmdx.EN, &DemoOptions{})
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range mustParsedFlags {
		flag := cmd.Flags().Lookup(name)
		NewWithT(t).Expect(flag).NotTo(BeNil())
	}

	for _, name := range mustNotParsedFlags {
		flag := cmd.Flags().Lookup(name)
		NewWithT(t).Expect(flag).To(BeNil())
	}

	t.Run("NameConflict", func(t *testing.T) {
		type Inline struct {
			Int int
		}

		flags := make(map[string]*cmdx.Flag)
		err = cmdx.ParseFlags(&struct {
			Int int
			Inline
		}{}, cmdx.EN, flags)
		NewWithT(t).Expect(err).NotTo(BeNil())
		NewWithT(t).Expect(err.Error()).To(ContainSubstring("name conflict"))
	})

	t.Run("NewCommandFailedToParseFlag", func(t *testing.T) {
		_, err = cmdx.NewCommand(cmdx.EN, InvalidExecutor{})
		NewWithT(t).Expect(err).NotTo(BeNil())
		NewWithT(t).Expect(err.Error()).To(Equal("expect value can be set"))
	})

	t.Run("CannotSet", func(t *testing.T) {
		flags := make(map[string]*cmdx.Flag)
		err = cmdx.ParseFlags(struct{}{}, cmdx.EN, flags)
		NewWithT(t).Expect(err).NotTo(BeNil())
		NewWithT(t).Expect(err.Error()).To(Equal("expect value can be set"))
	})

	t.Run("NotStruct", func(t *testing.T) {
		flags := make(map[string]*cmdx.Flag)
		err = cmdx.ParseFlags(ptrx.Ptr(100), cmdx.EN, flags)
		NewWithT(t).Expect(err).NotTo(BeNil())
		NewWithT(t).Expect(err.Error()).To(Equal("expect a struct value"))
	})

	flags := make(map[string]*cmdx.Flag)
	err = cmdx.ParseFlags(&DemoOptions{}, cmdx.EN, flags)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(flags["is-required"].Name())
	t.Log(flags["is-required"].LangHelp(cmdx.Lang("xx")))
	t.Log(flags["is-required"].LangHelp(cmdx.ZH))
	t.Log(flags["is-required"].LangHelp(cmdx.EN))
	t.Log(flags["is-required"].Env("prefix"))

	cmd, err = cmdx.NewCommand(cmdx.EN, &DemoOptions{})
	NewWithT(t).Expect(err).To(BeNil())

	flag := cmd.Flags().Lookup("is-required")
	flag.Value.Set("set require")
	flag.Changed = true
	NewWithT(t).Expect(cmd.Execute()).To(BeNil())
}

func (v *DemoOptions) Use() string { return "demo" }

func (v *DemoOptions) Short() string { return "demo short" }

func (v *DemoOptions) Exec(cmd *cobra.Command) error {
	cmd.Println(cmd.Flag("is-required").Value.String())
	return nil
}

type InvalidExecutor struct{}

func (v InvalidExecutor) Use() string { return "demo" }

func (v InvalidExecutor) Short() string { return "demo short" }

func (v InvalidExecutor) Exec(_ *cobra.Command) error { return nil }
