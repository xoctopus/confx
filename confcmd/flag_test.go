package confcmd_test

import (
	"reflect"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/xoctopus/confx/confcmd"
)

func TestNewFlagByStructField(t *testing.T) {
	some := &struct {
		Field1 int     `cmd:",required"         help.en:"field1 help" help.zh:"帮助1"`
		Field2 string  `cmd:"field"     env:"-" help.en:"field2 help" help.zh:"帮助2"`
		Field3 float64 `cmd:"-"`
		Field4 int64   `                env:"overwrite env"`
	}{
		Field2: "default value",
	}

	{
		sf := reflect.TypeOf(some).Elem().Field(0)
		fv := reflect.ValueOf(some).Elem().Field(0)

		flag := confcmd.NewFlagByStructInfo("prev", sf, fv)
		NewWithT(t).Expect(flag.Name()).To(Equal("prev-field-1"))
		NewWithT(t).Expect(flag.EnvKey("some")).To(Equal("SOME__PREV_FIELD_1"))
		NewWithT(t).Expect(flag.Help(confcmd.LangZH)).To(Equal("帮助1"))
		NewWithT(t).Expect(flag.Help(confcmd.LangEN)).To(Equal("field1 help"))
		NewWithT(t).Expect(flag.IsRequired()).To(BeTrue())
		_, ok := flag.Value().(int)
		NewWithT(t).Expect(ok).To(BeTrue())
		_, ok = flag.ValueVarP().(*int)
		NewWithT(t).Expect(ok).To(BeTrue())
		NewWithT(t).Expect(flag.DefaultValue()).To(Equal(some.Field1))
	}

	{
		sf := reflect.TypeOf(some).Elem().Field(1)
		fv := reflect.ValueOf(some).Elem().Field(1)

		flag := confcmd.NewFlagByStructInfo("", sf, fv)
		NewWithT(t).Expect(flag.Name()).To(Equal("field"))
		NewWithT(t).Expect(flag.EnvKey("some")).To(Equal(""))
		NewWithT(t).Expect(flag.Help(confcmd.LangZH)).To(Equal("帮助2"))
		NewWithT(t).Expect(flag.Help(confcmd.LangEN)).To(Equal("field2 help"))
		NewWithT(t).Expect(flag.IsRequired()).To(BeFalse())
		_, ok := flag.Value().(string)
		NewWithT(t).Expect(ok).To(BeTrue())
		_, ok = flag.ValueVarP().(*string)
		NewWithT(t).Expect(ok).To(BeTrue())
		NewWithT(t).Expect(flag.DefaultValue()).To(Equal(some.Field2))
	}

	{
		sf := reflect.TypeOf(some).Elem().Field(2)
		fv := reflect.ValueOf(some).Elem().Field(2)

		flag := confcmd.NewFlagByStructInfo("any", sf, fv)
		NewWithT(t).Expect(flag).To(BeNil())
	}

	{
		sf := reflect.TypeOf(some).Elem().Field(3)
		fv := reflect.ValueOf(some).Elem().Field(3)

		flag := confcmd.NewFlagByStructInfo("any", sf, fv)
		NewWithT(t).Expect(flag.Help(confcmd.LangZH)).To(Equal(""))
		NewWithT(t).Expect(flag.Help(confcmd.LangEN)).To(Equal(""))
		NewWithT(t).Expect(flag.EnvKey("")).To(Equal("ANY_OVERWRITE_ENV"))
	}
}
