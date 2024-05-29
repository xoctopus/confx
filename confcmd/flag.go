package confcmd

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/misc/stringsx"
	"github.com/xoctopus/x/ptrx"
	"github.com/xoctopus/x/reflectx"
)

func NewFlagByStructInfo(prefix string, sf reflect.StructField, fv reflect.Value) *Flag {
	field := strings.TrimPrefix(prefix+"."+sf.Name, ".")
	f := &Flag{
		field: field,
		name:  field,
		env:   ptrx.Ptr(field),
		helps: make(map[string]string),
	}

	if v, ok := sf.Tag.Lookup(FlagCmd); ok {
		tagKey, flags := reflectx.ParseTagKeyAndFlags(v)
		if tagKey == "-" {
			return nil
		}
		if tagKey != "" {
			f.name = strings.TrimPrefix(prefix+"."+tagKey, ".")
		}
		if _, ok = flags[FlagRequired]; ok {
			f.required = true
		}
	}

	if v, ok := sf.Tag.Lookup(FlagEnv); ok {
		tagKey, _ := reflectx.ParseTagKeyAndFlags(v)
		if tagKey == "-" {
			f.env = nil
		} else {
			if tagKey != "" {
				*f.env = strings.TrimPrefix(prefix+"."+tagKey, ".")
			}
		}
	}

	for _, key := range MultiLangHelpKeys {
		if tagKey, ok := sf.Tag.Lookup(key); ok {
			help, _ := reflectx.ParseTagKeyAndFlags(tagKey)
			f.helps[key] = help
		}
	}

	f.name = stringsx.LowerDashJoint(f.name)
	if f.env != nil {
		*f.env = stringsx.UpperSnakeCase(*f.env)
	}

	f.value = reflectx.IndirectNew(fv)
	f.defaults = f.value.Interface()

	return f
}

type Flag struct {
	// field struct field name
	field string
	// name command flag name, eg: some-flag
	name string
	// env key can injector from env var, eg: SOME_FLAG
	env *string
	// helps multi-language help info
	helps map[string]string
	// required if this flag is required
	required bool
	// defaults flag's default value
	defaults any
	// value filed value
	value reflect.Value
}

func (f *Flag) Name() string {
	return f.name
}

func (f *Flag) Field() string {
	return f.field
}

func (f *Flag) Help(lang LangType) string {
	keys := make([]string, 0, len(MultiLangHelpKeys)+1)
	keys = append(keys, strings.ToLower(lang.HelpKey()))
	keys = append(keys, MultiLangHelpKeys...)

	for _, name := range keys {
		if help, _ := f.helps[name]; help != "" {
			return help
		}
	}
	return ""
}

func (f *Flag) IsRequired() bool {
	return f.required
}

func (f *Flag) DefaultValue() any {
	return f.defaults
}

func (f *Flag) Value() any {
	must.BeTrueWrap(
		f.value.IsValid() && f.value.CanInterface(),
		"field is not valid or cannot interface: %s", f.name,
	)
	return f.value.Interface()
}

func (f *Flag) ValueVarP() any {
	must.BeTrueWrap(
		f.value.IsValid() && f.value.CanAddr() && f.value.Addr().CanInterface(),
		"field is not valid or cannot interface, or addr cannot interface: %s", f.name,
	)
	return f.value.Addr().Interface()
}

func (f *Flag) EnvKey(prefix string) string {
	if f.env == nil {
		return ""
	}
	if prefix != "" {
		return strings.Replace(strings.ToUpper(prefix+"__"+*f.env), "-", "_", -1)
	}
	return strings.Replace(strings.ToUpper(*f.env), "-", "_", -1)
}

func (f *Flag) Register(cmd *cobra.Command, lang LangType) error {
	flags := cmd.Flags()
	help := f.Help(lang)
	switch v := f.defaults.(type) {
	case bool:
		flags.BoolVarP(f.ValueVarP().(*bool), f.name, "", v, help)
	case string:
		flags.StringVarP(f.ValueVarP().(*string), f.name, "", v, help)
	case int:
		flags.IntVarP(f.ValueVarP().(*int), f.name, "", v, help)
	case int8:
		flags.Int8VarP(f.ValueVarP().(*int8), f.name, "", v, help)
	case int16:
		flags.Int16VarP(f.ValueVarP().(*int16), f.name, "", v, help)
	case int32:
		flags.Int32VarP(f.ValueVarP().(*int32), f.name, "", v, help)
	case int64:
		flags.Int64VarP(f.ValueVarP().(*int64), f.name, "", v, help)
	case uint:
		flags.UintVarP(f.ValueVarP().(*uint), f.name, "", v, help)
	case uint8:
		flags.Uint8VarP(f.ValueVarP().(*uint8), f.name, "", v, help)
	case uint16:
		flags.Uint16VarP(f.ValueVarP().(*uint16), f.name, "", v, help)
	case uint32:
		flags.Uint32VarP(f.ValueVarP().(*uint32), f.name, "", v, help)
	case uint64:
		flags.Uint64VarP(f.ValueVarP().(*uint64), f.name, "", v, help)
	case float32:
		flags.Float32VarP(f.ValueVarP().(*float32), f.name, "", v, help)
	case float64:
		flags.Float64VarP(f.ValueVarP().(*float64), f.name, "", v, help)
	case []int:
		flags.IntSliceVarP(f.ValueVarP().(*[]int), f.name, "", v, help)
	case []int32:
		flags.Int32SliceVarP(f.ValueVarP().(*[]int32), f.name, "", v, help)
	case []int64:
		flags.Int64SliceVarP(f.ValueVarP().(*[]int64), f.name, "", v, help)
	case []uint:
		flags.UintSliceVarP(f.ValueVarP().(*[]uint), f.name, "", v, help)
	case []float32:
		flags.Float32SliceVarP(f.ValueVarP().(*[]float32), f.name, "", v, help)
	case []float64:
		flags.Float64SliceVarP(f.ValueVarP().(*[]float64), f.name, "", v, help)
	case []string:
		flags.StringSliceVarP(f.ValueVarP().(*[]string), f.name, "", v, help)
	case []bool:
		flags.BoolSliceVarP(f.ValueVarP().(*[]bool), f.name, "", v, help)
	default:
		return errors.Errorf("unsupported flag value type: `%s`", f.value.Type())
	}
	if f.required {
		return cmd.MarkFlagRequired(f.name)
	}
	return nil
}
