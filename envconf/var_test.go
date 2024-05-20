package envconf_test

import (
	"fmt"
	"os"

	"github.com/sincospro/conf/envconf"
)

func ExampleParseGroupFromEnv() {
	_ = os.Setenv("TEST__Array_0_A", "string")
	_ = os.Setenv("TEST__Array_0_B", "2")
	_ = os.Setenv("TEST__MapStringInt64_Key", "101")
	_ = os.Setenv("TEST__MapStringInt_Key", "100")
	_ = os.Setenv("TEST__MapStringPassword_Key", "password")
	_ = os.Setenv("TEST__MapStringString_Key", "Value")
	_ = os.Setenv("TEST__Slice_0_A", "2")
	_ = os.Setenv("TEST__Slice_0_B", "string")

	grp := envconf.ParseGroupFromEnv("TEST")

	_ = os.Unsetenv("TEST__Array_0_A")
	_ = os.Unsetenv("TEST__Array_0_B")
	_ = os.Unsetenv("TEST__MapStringInt64_Key")
	_ = os.Unsetenv("TEST__MapStringInt_Key")
	_ = os.Unsetenv("TEST__MapStringPassword_Key")
	_ = os.Unsetenv("TEST__MapStringString_Key")
	_ = os.Unsetenv("TEST__Slice_0_A")
	_ = os.Unsetenv("TEST__Slice_0_B")

	fmt.Print(string(grp.DotEnv(nil)))

	// output:
	// TEST__Array_0_A=string
	// TEST__Array_0_B=2
	// TEST__MapStringInt64_Key=101
	// TEST__MapStringInt_Key=100
	// TEST__MapStringPassword_Key=password
	// TEST__MapStringString_Key=Value
	// TEST__Slice_0_A=2
	// TEST__Slice_0_B=string
}
