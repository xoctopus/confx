package confcmd

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/xoctopus/x/reflectx"
)

func NewCommand(lang Lang, v Executor) (*cobra.Command, error) {
	flags := make(map[string]*Flag)

	if err := ParseFlags(v, lang, flags); err != nil {
		return nil, err
	}

	cmd := &cobra.Command{
		Use:   v.Use(),
		Short: v.Short(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Exec(cmd)
		},
	}

	for _, f := range flags {
		if err := f.Register(cmd); err != nil {
			continue
		}
	}

	return cmd, nil
}

func ParseFlags(v any, lang Lang, flags map[string]*Flag, prefixes ...string) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflectx.Indirect(reflect.ValueOf(v))
	}

	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() && rv.CanSet() {
			rv = reflectx.IndirectNew(rv)
		}
		return ParseFlags(rv, lang, flags, prefixes...)
	}

	if rv.Kind() != reflect.Struct {
		return errors.Errorf("expect a struct value")
	}

	if !rv.CanSet() {
		return errors.Errorf("expect value can be set")
	}

	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		frt := rt.Field(i)

		if !frt.IsExported() {
			continue
		}

		prefix := strings.Join(prefixes, "-")
		f := NewFlagByStructField(prefix, lang, frt)
		if f == nil {
			continue
		}

		if _, ok = flags[f.name]; ok {
			return errors.Errorf("name conflict: [flag: %s] [field: %s]", f.name, frt.Name)
		}

		frv := rv.Field(i)
		if reflectx.Deref(frt.Type).Kind() == reflect.Struct {
			tagVal, _ := frt.Tag.Lookup(flagKey)
			_, tagFlags := reflectx.ParseTagKeyAndFlags(tagVal)
			_prefixes := prefixes
			if !(len(tagFlags) == 0 && frt.Anonymous) {
				_prefixes = append(_prefixes, f.name)
			}
			if err := ParseFlags(frv, lang, flags, _prefixes...); err != nil {
				return err
			}
			continue
		}
		f.value = reflectx.IndirectNew(frv)
		f.defaults = f.value.Interface()
		flags[f.name] = f
	}
	return nil
}
