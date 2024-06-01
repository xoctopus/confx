package confcmd

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/textx"
)

func NewFlagValue(rv reflect.Value) pflag.Value {
	return &flagValue{rv: rv}
}

// flagValue refers the Flag.value, implements pflag.Value
// it can receive all basic types, such as `int`, `uint`. and types that
// implement encoding.Unmarshaller/Marshaller.
// it is also acceptable if they are elements of a one-dimensional slice.
// but multidimensional slices are not supported.
type flagValue struct {
	rv reflect.Value
}

func (fv *flagValue) String() string {
	if fv.rv.Kind() == reflect.Slice {
		fields := make([]string, fv.rv.Len())
		for i := 0; i < fv.rv.Len(); i++ {
			raw, err := textx.MarshalText(fv.rv.Index(i))
			must.NoErrorWrap(err, "failed to marshal %s at index %d", fv.Type(), i)
			fields[i] = string(raw)
		}
		return "[" + strings.Join(fields, " ") + "]"
	}

	raw, err := textx.MarshalText(fv.rv)
	must.NoErrorWrap(err, "failed to marshal %s", fv.Type())

	return string(raw)
}

func (fv *flagValue) Set(val string) error {
	if val == "" {
		return nil
	}
	var err error

	if fv.rv.Kind() == reflect.Slice {
		fields := strings.Fields(val)
		for i, field := range fields {
			v := reflect.New(fv.rv.Type().Elem()).Elem()
			if err = textx.UnmarshalText([]byte(field), v); err != nil {
				return errors.Wrapf(
					err, "failed to parse %s at index %d by %s",
					fv.Type(), i, val,
				)
			}
			fv.rv.Set(reflect.Append(fv.rv, v))
		}
	} else {
		if err = textx.UnmarshalText([]byte(val), fv.rv); err != nil {
			err = errors.Wrapf(err, "failed to parse %s from %s", fv.Type(), val)
		}
	}

	return err
}

func (fv *flagValue) Type() string {
	return fv.rv.Type().String()
}
