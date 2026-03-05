package helper_test

import (
	"testing"

	. "cgtech.gitlab.com/saitox/x/testx"

	"cgtech.gitlab.com/saitox/confx/pkg/helper"
)

func TestHostIdentifier(t *testing.T) {
	Expect(t, helper.HostIdentifier(""), Equal(helper.DefaultHostIdentifier()))
	Expect(t, helper.HostIdentifier("AppName"), HavePrefix("AppName"))
}
