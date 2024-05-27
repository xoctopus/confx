package envconf_test

import (
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/xoctopus/confx/envconf"
)

func TestErrUnexpectMapKeyType(t *testing.T) {
	var (
		err            = envconf.NewErrUnmarshalUnexpectMapKeyType(reflect.TypeOf(100.0))
		targetEqual    = envconf.NewErrUnmarshalUnexpectMapKeyType(reflect.TypeOf(101.0))
		targetNotEqual = envconf.NewErrUnmarshalUnexpectMapKeyType(nil)
	)

	NewWithT(t).Expect(errors.Is(err, targetNotEqual)).To(BeFalse())
	NewWithT(t).Expect(errors.Is(err, targetEqual)).To(BeTrue())
	NewWithT(t).Expect(errors.Is(err, errors.New("any"))).To(BeFalse())

	t.Log(err.Error())
	t.Log(targetNotEqual)
	t.Log(envconf.ErrUnexpectMapKeyValue("so-me"))
}

func TestErrInvalidDecodeValue(t *testing.T) {
	var (
		err            = envconf.NewInvalidValueErr()
		targetEqual    = envconf.NewInvalidDecodeError(nil, "invalid value")
		targetNotEqual = envconf.NewCannotSetErr(reflect.TypeOf(1))
	)

	NewWithT(t).Expect(errors.Is(err, targetEqual)).To(BeTrue())
	NewWithT(t).Expect(errors.Is(err, targetNotEqual)).To(BeFalse())
	NewWithT(t).Expect(errors.Is(err, errors.New("any"))).To(BeFalse())

	t.Log(err.Error())
	t.Log(targetNotEqual)
}
