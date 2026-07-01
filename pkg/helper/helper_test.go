package helper_test

import (
	"testing"
	"time"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/helper"
)

func TestHostIdentifier(t *testing.T) {
	Expect(t, helper.HostIdentifier(""), Equal(helper.DefaultHostIdentifier()))
	Expect(t, helper.HostIdentifier("AppName"), HavePrefix("AppName"))
}

func TestCost(t *testing.T) {
	du := time.Millisecond * 100
	span := helper.Span()
	time.Sleep(du)
	Expect(t, span() >= du, BeTrue())
}
