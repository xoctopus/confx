package confcmd

import (
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/reflectx"
)

func NewCommand(v Executor) *cobra.Command {
	flags := ParseFlags(v)

	cmd := &cobra.Command{
		Use:   v.Use(),
		Short: v.Short(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Exec(cmd, args...)
		},
	}

	if executor, ok := v.(WithLong); ok {
		cmd.Long = executor.Long()
	}
	if executor, ok := v.(WithExample); ok {
		cmd.Example = executor.Example()
	}

	lang := v.HelpLang()
	injector, _ := v.(CanInjectFromEnv)
	for _, f := range flags {
		v.AddFlag(f)
		envVar := ""
		if injector != nil {
			envVar = os.Getenv(f.EnvKey(injector.Prefix()))
		}
		err := f.Register(cmd, lang, envVar)
		must.NoErrorWrap(err, "failed to registered flag: %s", f.name)
	}

	return cmd
}

func ParseFlags(v any) []*Flag {
	return parseFlags(v, "")
}

func parseFlags(v any, prefix string) []*Flag {
	flags := make([]*Flag, 0)

	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflectx.Indirect(reflect.ValueOf(v))
	}

	must.BeTrueWrap(rv != reflectx.InvalidValue, "invalid input value")

	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() && rv.CanSet() {
			rv = reflectx.IndirectNew(rv)
		} else {
			rv = rv.Elem()
		}
		flags = append(flags, parseFlags(rv, prefix)...)
		return flags
	}

	rt := rv.Type()

	must.BeTrueWrap(rv.Kind() == reflect.Struct, "expect a struct value, but got %s")

	must.BeTrueWrap(
		rv.CanSet(),
		"expect value can set: [type: %s] [prefix: %s]", rt, prefix,
	)

	for i := 0; i < rv.NumField(); i++ {
		sf := rv.Type().Field(i)
		if !sf.IsExported() {
			continue
		}

		name := sf.Name
		tagKey, _ := reflectx.ParseTagKeyAndFlags(sf.Tag.Get(FlagCmd))
		if tagKey == "-" {
			continue
		}
		if tagKey != "" {
			name = tagKey
		}

		fv := rv.Field(i)
		if reflectx.Deref(fv.Type()).Kind() == reflect.Struct {
			_prefix := strings.TrimPrefix(prefix+"."+name, ".")
			if sf.Anonymous && sf.Tag == "" {
				_prefix = prefix
			}
			flags = append(flags, parseFlags(fv, _prefix)...)
			continue
		}
		if flag := NewFlagByStructInfo(prefix, sf, fv); flag != nil {
			flags = append(flags, flag)
		}
	}
	return flags
}
