package helper_test

import (
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/helper"
)

func TestHostIdentifier(t *testing.T) {
	Expect(t, helper.HostIdentifier(""), Equal(helper.DefaultHostIdentifier()))
	Expect(t, helper.HostIdentifier("AppName"), HavePrefix("AppName"))
}
