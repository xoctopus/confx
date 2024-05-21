package confcmd_test

import (
	"reflect"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/sincospro/conf/confcmd"
)

func TestNewFlagByStructField(t *testing.T) {
	some := struct {
		Field int `name:",required" help.en:"field help" help.zh:"帮助"`
	}{}

	flag := confcmd.NewFlagByStructField("prefix", confcmd.ZH, reflect.TypeOf(some).Field(0))
	NewWithT(t).Expect(flag.Name()).To(Equal("prefix-field"))
	NewWithT(t).Expect(flag.Help()).To(Equal("帮助"))
	NewWithT(t).Expect(flag.IsRequired()).To(BeTrue())
}
