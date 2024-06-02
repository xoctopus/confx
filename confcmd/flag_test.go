package confcmd_test

import (
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	. "github.com/xoctopus/confx/confcmd"
)

func TestNewFlagByStructField(t *testing.T) {
	some := &struct {
		Field1 int     `cmd:",required"         help.en:"field1 help" help.zh:"帮助1"`
		Field2 string  `cmd:"field,r"   env:"-" help.en:"field2 help" help.zh:"帮助2"`
		Field3 float64 `cmd:"-"`
		Field4 int64   `cmd:",p"        env:"overwrite env"`
		Field5 uint    `cmd:",persis"`
		Field6 bool    `cmd:",nop=1"`
		Field7 string  `cmd:",s=s"`
	}{
		Field2: "default value",
	}

	{
		sf := reflect.TypeOf(some).Elem().Field(0)
		fv := reflect.ValueOf(some).Elem().Field(0)

		flag := NewFlagByStructInfo("prev", sf, fv)
		NewWithT(t).Expect(flag.Name()).To(Equal("prev-field-1"))
		NewWithT(t).Expect(flag.EnvKey("some")).To(Equal("SOME__PREV_FIELD_1"))
		NewWithT(t).Expect(flag.Help(LangZH)).To(Equal("帮助1"))
		NewWithT(t).Expect(flag.Help(LangEN)).To(Equal("field1 help"))
		NewWithT(t).Expect(flag.IsRequired()).To(BeTrue())
		_, ok := flag.Value().(int)
		NewWithT(t).Expect(ok).To(BeTrue())
		_, ok = flag.ValueVarP().(*int)
		NewWithT(t).Expect(ok).To(BeTrue())
		NewWithT(t).Expect(flag.DefaultValue()).To(Equal(some.Field1))

		err := flag.Register(&cobra.Command{}, LangZH, "100")
		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(flag.Value()).To(Equal(100))
	}

	{
		sf := reflect.TypeOf(some).Elem().Field(1)
		fv := reflect.ValueOf(some).Elem().Field(1)

		flag := NewFlagByStructInfo("", sf, fv)
		NewWithT(t).Expect(flag.Name()).To(Equal("field"))
		NewWithT(t).Expect(flag.EnvKey("some")).To(Equal(""))
		NewWithT(t).Expect(flag.Help(LangZH)).To(Equal("帮助2"))
		NewWithT(t).Expect(flag.Help(LangEN)).To(Equal("field2 help"))
		_, ok := flag.Value().(string)
		NewWithT(t).Expect(ok).To(BeTrue())
		_, ok = flag.ValueVarP().(*string)
		NewWithT(t).Expect(ok).To(BeTrue())
		NewWithT(t).Expect(flag.DefaultValue()).To(Equal(some.Field2))

		err := flag.Register(&cobra.Command{}, LangZH, "some string")
		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(flag.Value()).To(Equal("some string"))

		NewWithT(t).Expect(flag.IsRequired()).To(BeTrue())
		NewWithT(t).Expect(flag.IsPersistent()).To(BeFalse())
	}

	{
		sf := reflect.TypeOf(some).Elem().Field(2)
		fv := reflect.ValueOf(some).Elem().Field(2)

		flag := NewFlagByStructInfo("any", sf, fv)
		NewWithT(t).Expect(flag).To(BeNil())
	}

	{
		sf := reflect.TypeOf(some).Elem().Field(3)
		fv := reflect.ValueOf(some).Elem().Field(3)

		flag := NewFlagByStructInfo("any", sf, fv)
		NewWithT(t).Expect(flag.Help(LangZH)).To(Equal(""))
		NewWithT(t).Expect(flag.Help(LangEN)).To(Equal(""))
		NewWithT(t).Expect(flag.EnvKey("")).To(Equal("ANY_OVERWRITE_ENV"))

		err := flag.Register(&cobra.Command{}, LangZH, "invalid value")
		NewWithT(t).Expect(err).NotTo(BeNil())
		NewWithT(t).Expect(err.Error()).To(ContainSubstring("failed to parse int64"))

		NewWithT(t).Expect(flag.IsRequired()).To(BeFalse())
		NewWithT(t).Expect(flag.IsPersistent()).To(BeTrue())
	}

	{
		sf := reflect.TypeOf(some).Elem().Field(4)
		fv := reflect.ValueOf(some).Elem().Field(4)

		flag := NewFlagByStructInfo("", sf, fv)
		NewWithT(t).Expect(flag.IsPersistent()).To(BeTrue())
		NewWithT(t).Expect(flag.NoOptionDefaultValue()).To(Equal(""))
	}

	{
		sf := reflect.TypeOf(some).Elem().Field(5)
		fv := reflect.ValueOf(some).Elem().Field(5)

		flag := NewFlagByStructInfo("", sf, fv)
		NewWithT(t).Expect(flag.NoOptionDefaultValue()).To(Equal("1"))
	}

	{
		sf := reflect.TypeOf(some).Elem().Field(6)
		fv := reflect.ValueOf(some).Elem().Field(6)

		flag := NewFlagByStructInfo("", sf, fv)
		NewWithT(t).Expect(flag.Shorthand()).To(Equal("s"))
		err := flag.Register(&cobra.Command{}, LangZH, "")
		NewWithT(t).Expect(err).To(BeNil())
	}

}
