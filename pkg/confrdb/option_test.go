package confrdb_test

import (
	"testing"

	. "github.com/xoctopus/x/testx"
	"github.com/xoctopus/x/textx"

	"github.com/xoctopus/confx/pkg/confrdb"
	"github.com/xoctopus/confx/pkg/confrdb/option"
)

func TestOption(t *testing.T) {
	o1 := &confrdb.Option[option.MySQL]{}
	o1.SetDefault()
	uv1, err := textx.MarshalURL(o1)
	Expect(t, err, Succeed())

	mo := &option.MySQL{}
	mo.SetDefault()
	uv2, err := textx.MarshalURL(o1)
	Expect(t, err, Succeed())

	Expect(t, uv1.Encode(), Equal(uv2.Encode()))
}
