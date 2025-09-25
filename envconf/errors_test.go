package envconf_test

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"github.com/xoctopus/x/misc/must"

	. "github.com/xoctopus/confx/envconf"
)

func ExampleError() {
	pos := NewPathWalker()
	pos.Enter("ENTRY")

	// assertion: wrap a nil error, result must be nil
	must.BeTrue(NewErrorW(pos, E_DEC__FAILED_UNMARSHAL, nil) == nil)

	cases := []error{
		NewError(pos, E_DEC__INVALID_VALUE),
		NewError(pos, E_DEC__INVALID_VALUE_CANNOT_SET),
		NewErrorf(pos, E_DEC__INVALID_MAP_KEY_TYPE, "`%s`", reflect.TypeOf(1)),
		NewErrorW(pos, E_DEC__FAILED_UNMARSHAL, errors.New("UNMARSHAL FAILED")),
		NewErrorf(pos, E_ENC__INVALID_MAP_KEY_TYPE, "`%s`", reflect.TypeOf([3]int{})),
		NewErrorf(pos, E_ENC__INVALID_MAP_KEY_VALUE, "`%d`", 100),
		NewErrorW(pos, E_ENC__FAILED_MARSHAL, errors.New("MARSHAL FAILED")),
	}
	for _, e := range cases {
		fmt.Println(e.Error())
	}

	// Output:
	// when decoding at `ENTRY`, got cannot set value
	// when decoding at `ENTRY`, got nil value
	// when decoding at `ENTRY`, got invalid map key type, expect alphabet string or positive integer. but got: `int`
	// when decoding at `ENTRY`, failed to unmarshal [UNMARSHAL FAILED]
	// when encoding at `ENTRY`, got invalid map key type, expect alphabet string or positive integer. but got: `[3]int`
	// when encoding at `ENTRY`, got invalid map key value, expect alphabet string or positive integer. but got: `100`
	// when encoding at `ENTRY`, failed to marshal. [MARSHAL FAILED]
}
