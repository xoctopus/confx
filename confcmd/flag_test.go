package confcmd_test

import (
	"math/big"
	"os"
	"testing"

	"github.com/jedib0t/go-pretty/v6/table"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"github.com/xoctopus/x/reflectx"

	. "github.com/xoctopus/confx/confcmd"
)

type Embed struct {
	Require      string   `cmd:",require" env:"-"`                                     // required and cannot inject by env
	Persistent   *float32 `cmd:",persist"`                                             // persistent flag
	NoOptDefault bool     `cmd:",default=true"`                                        // no option default true
	HasDefault   int      `cmd:"" env:"HAS_DEFAULT_INT"`                               // rewrite env key
	Shorthand    *int     `cmd:"ss,short=s" env:""`                                    // rewrite flag name. flag name is ss, shorthand is s
	HasI18nID    *string  `cmd:"has-i18n-id" env:"HAS_I18N_ID" i18n:"embed.HasI18nID"` // has i18n message id
	HasHelpUsage string   `cmd:"" help:"help usage"`                                   // has help usage
}

type Complex struct {
	Strings  []*string `cmd:""`        // string pointer slice
	Integers []int64   `cmd:""`        // int slice
	BigInt   *big.Int  `cmd:",inline"` // single flag
}

type TestFlags struct {
	unexported string
	Ignored    int    `cmd:"-"`
	EmbedPtr   *Embed `cmd:"group"`
	UnameEmbed Embed
	*Embed
	Complex
}

func localizer(i18nId string) string {
	if i18nId == "embed.HasI18nID" {
		return "帮助"
	}
	return ""
}

func TestParseFlags(t *testing.T) {
	t.Run("InvalidInputValue", func(t *testing.T) {
		for _, v := range []any{nil, (*struct{})(nil), reflectx.InvalidValue} {
			t.Run("", func(t *testing.T) {
				defer func() {
					e, ok := recover().(error)
					NewWithT(t).Expect(ok).To(BeTrue())
					NewWithT(t).Expect(e).NotTo(BeNil())
					NewWithT(t).Expect(e.Error()).To(ContainSubstring("invalid input value"))
				}()
				ParseFlags(v)
			})
		}
	})
	t.Run("NotStruct", func(t *testing.T) {
		defer func() {
			e, ok := recover().(error)
			NewWithT(t).Expect(ok).To(BeTrue())
			NewWithT(t).Expect(e).NotTo(BeNil())
			NewWithT(t).Expect(e.Error()).To(ContainSubstring("expect a struct value"))
		}()
		ParseFlags(1)
	})
	t.Run("CannotSet", func(t *testing.T) {
		defer func() {
			e, ok := recover().(error)
			NewWithT(t).Expect(ok).To(BeTrue())
			NewWithT(t).Expect(e).NotTo(BeNil())
			NewWithT(t).Expect(e.Error()).To(ContainSubstring("expect value can set"))
		}()
		ParseFlags(struct{}{})
	})
	t.Run("ParseFlags", func(t *testing.T) {
		v := &TestFlags{Embed: &Embed{HasDefault: 100}}

		flags := ParseFlags(v)
		flagMap := make(map[string]*Flag)
		for _, f := range flags {
			flagMap[f.Name()] = f
		}

		t.Run("CheckParseResult", func(t *testing.T) {
			t.Run("Require", func(t *testing.T) {
				f := flagMap["require"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.IsRequired()).To(BeTrue())
				NewWithT(t).Expect(f.Env()).To(Equal(""))
				NewWithT(t).Expect(f.NoOptionDefaultValue()).To(Equal(""))

				f = flagMap["group-require"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.IsRequired()).To(BeTrue())
				NewWithT(t).Expect(f.Env()).To(Equal(""))
				NewWithT(t).Expect(f.NoOptionDefaultValue()).To(Equal(""))

				f = flagMap["uname-embed-require"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.IsRequired()).To(BeTrue())
				NewWithT(t).Expect(f.Env()).To(Equal(""))
				NewWithT(t).Expect(f.NoOptionDefaultValue()).To(Equal(""))
			})
			t.Run("Persistent", func(t *testing.T) {
				f := flagMap["persistent"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.IsPersistent()).To(BeTrue())

				f = flagMap["group-persistent"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.IsPersistent()).To(BeTrue())

				f = flagMap["uname-embed-persistent"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.IsPersistent()).To(BeTrue())
			})
			t.Run("NoOptDefValue", func(t *testing.T) {
				f := flagMap["no-opt-default"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.NoOptionDefaultValue()).To(Equal("true"))

				f = flagMap["group-no-opt-default"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.NoOptionDefaultValue()).To(Equal("true"))

				f = flagMap["uname-embed-no-opt-default"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.NoOptionDefaultValue()).To(Equal("true"))
			})
			t.Run("HasDefault", func(t *testing.T) {
				f := flagMap["has-default"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.Env()).To(Equal("HAS_DEFAULT_INT"))

				f = flagMap["group-has-default"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.Env()).To(Equal("GROUP_HAS_DEFAULT_INT"))

				f = flagMap["uname-embed-has-default"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.Env()).To(Equal("UNAME_EMBED_HAS_DEFAULT_INT"))
			})
			t.Run("ShortHand", func(t *testing.T) {
				NewWithT(t).Expect(flagMap["short-hand"]).To(BeNil())

				f := flagMap["ss"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.Name()).To(Equal("ss"))
				NewWithT(t).Expect(f.Short()).To(Equal("s"))

				f = flagMap["group-ss"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.Name()).To(Equal("group-ss"))
				NewWithT(t).Expect(f.Short()).To(Equal("s"))

				f = flagMap["uname-embed-ss"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.Name()).To(Equal("uname-embed-ss"))
				NewWithT(t).Expect(f.Short()).To(Equal("s"))
			})
			t.Run("HasI18nMsgID", func(t *testing.T) {
				f := flagMap["has-i18n-id"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.I18n()).To(Equal("embed.HasI18nID"))

				f = flagMap["group-has-i18n-id"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.I18n()).To(Equal("embed.HasI18nID"))

				f = flagMap["uname-embed-has-i18n-id"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.I18n()).To(Equal("embed.HasI18nID"))
			})
			t.Run("HasHelpUsage", func(t *testing.T) {
				f := flagMap["has-help-usage"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.Help()).To(Equal("help usage"))

				f = flagMap["group-has-help-usage"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.Help()).To(Equal("help usage"))

				f = flagMap["uname-embed-has-help-usage"]
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.Help()).To(Equal("help usage"))
			})
		})

		t.Run("CheckRegistered", func(t *testing.T) {
			cmd := &cobra.Command{}

			t.Run("Require", func(t *testing.T) {
				flagMap["require"].Register(cmd, "abc", nil)
				f := cmd.Flag("require")
				NewWithT(t).Expect(f.Name).To(Equal("require"))
				NewWithT(t).Expect(f.DefValue).To(Equal("abc"))
			})
			t.Run("HasI18nMsgID", func(t *testing.T) {
				flagMap["has-i18n-id"].Register(cmd, "def", localizer)
				f := cmd.Flag("has-i18n-id")
				NewWithT(t).Expect(f.Name).To(Equal("has-i18n-id"))
				NewWithT(t).Expect(f.DefValue).To(Equal("def"))
				NewWithT(t).Expect(f.Usage).To(Equal("帮助"))
			})
			t.Run("Shorthand", func(t *testing.T) {
				flagMap["ss"].Register(cmd, "", nil)
				f := cmd.Flag("ss")
				NewWithT(t).Expect(f.Name).To(Equal("ss"))
				NewWithT(t).Expect(f.Shorthand).To(Equal("s"))
			})
			t.Run("PersistentFlag", func(t *testing.T) {
				flagMap["persistent"].Register(cmd, "", nil)
				f := cmd.Flag("persistent")
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.Name).To(Equal("persistent"))
				f = cmd.PersistentFlags().Lookup("persistent")
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.Name).To(Equal("persistent"))
			})
			t.Run("NoOptionDefault", func(t *testing.T) {
				flagMap["no-opt-default"].Register(cmd, "", nil)
				f := cmd.Flag("no-opt-default")
				NewWithT(t).Expect(f).NotTo(BeNil())
				NewWithT(t).Expect(f.NoOptDefVal).To(Equal("true"))
			})
			t.Run("FailedToSetValueFromEnv", func(t *testing.T) {
				t.Run("Integers", func(t *testing.T) {
					f := flagMap["integers"]
					NewWithT(t).Expect(f).NotTo(BeNil())

					defer func() {
						err, _ := recover().(error)
						NewWithT(t).Expect(err).NotTo(BeNil())
						NewWithT(t).Expect(err.Error()).To(ContainSubstring("failed to set env var"))
					}()
					f.Register(cmd, "abc", nil)
				})
				t.Run("BigInt", func(t *testing.T) {
					f := flagMap["big-int"]
					NewWithT(t).Expect(f).NotTo(BeNil())

					defer func() {
						err, _ := recover().(error)
						NewWithT(t).Expect(err).NotTo(BeNil())
						NewWithT(t).Expect(err.Error()).To(ContainSubstring("failed to set env var"))
					}()
					f.Register(cmd, "abc", nil)
				})
			})
		})
	})
}

func ExampleParseFlags() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "name", "short", "type", "env key", "required", "persistent", "help", "i18n id", "default", "no opt default"})

	v := &struct {
		*Embed
		Complex
	}{
		Complex: Complex{
			Strings: []*string{},
			BigInt:  big.NewInt(100),
		},
	}
	cmd := &cobra.Command{}
	flags := ParseFlags(v)

	_ = os.Setenv("STRINGS", "a b c")

	rows := make([]table.Row, 0, len(flags))
	for i, f := range flags {
		f.Register(cmd, os.Getenv(f.Env()), nil)
		ff := cmd.Flag(f.Name())
		row := table.Row{
			i + 1, ff.Name, ff.Shorthand, ff.Value.Type(), f.Env(), f.IsRequired(), f.IsPersistent(),
			f.Help(), f.I18n(), ff.DefValue, ff.NoOptDefVal,
		}
		rows = append(rows, row)
	}
	t.AppendRows(rows)
	t.SetStyle(table.StyleLight)
	t.Render()

	// Output:
	// ┌────┬────────────────┬───────┬───────────┬─────────────────┬──────────┬────────────┬────────────┬─────────────────┬─────────┬────────────────┐
	// │  # │ NAME           │ SHORT │ TYPE      │ ENV KEY         │ REQUIRED │ PERSISTENT │ HELP       │ I18N ID         │ DEFAULT │ NO OPT DEFAULT │
	// ├────┼────────────────┼───────┼───────────┼─────────────────┼──────────┼────────────┼────────────┼─────────────────┼─────────┼────────────────┤
	// │  1 │ big-int        │       │ big.Int   │ BIG_INT         │ false    │ false      │            │                 │ 100     │                │
	// │  2 │ has-default    │       │ int       │ HAS_DEFAULT_INT │ false    │ false      │            │                 │ 0       │                │
	// │  3 │ has-help-usage │       │ string    │ HAS_HELP_USAGE  │ false    │ false      │ help usage │                 │         │                │
	// │  4 │ has-i18n-id    │       │ string    │ HAS_I18N_ID     │ false    │ false      │            │ embed.HasI18nID │         │                │
	// │  5 │ integers       │       │ []int64   │ INTEGERS        │ false    │ false      │            │                 │ []      │                │
	// │  6 │ no-opt-default │       │ bool      │ NO_OPT_DEFAULT  │ false    │ false      │            │                 │ false   │ true           │
	// │  7 │ persistent     │       │ float32   │ PERSISTENT      │ false    │ true       │            │                 │ 0       │                │
	// │  8 │ require        │       │ string    │                 │ true     │ false      │            │                 │         │                │
	// │  9 │ ss             │ s     │ int       │ SHORTHAND       │ false    │ false      │            │                 │ 0       │                │
	// │ 10 │ strings        │       │ []*string │ STRINGS         │ false    │ false      │            │                 │ [a b c] │                │
	// └────┴────────────────┴───────┴───────────┴─────────────────┴──────────┴────────────┴────────────┴─────────────────┴─────────┴────────────────┘
}
