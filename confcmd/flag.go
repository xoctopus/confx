package confcmd

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/sincospro/x/misc/stringsx"
	"github.com/sincospro/x/reflectx"
	"github.com/spf13/cobra"
)

func NewFlag(name string, lang Lang) *Flag {
	return &Flag{
		name:     name,
		lang:     lang,
		helps:    make(map[string]string),
		required: false,
	}
}

func NewFlagByStructField(prefix string, lang Lang, sf reflect.StructField) *Flag {
	name := sf.Name
	if prefix != "" {
		name = prefix + "-" + name
	}
	f := NewFlag(name, lang)

	if v, ok := sf.Tag.Lookup(flagKey); ok {
		tagKey, flags := reflectx.ParseTagKeyAndFlags(v)
		if tagKey == "-" {
			return nil
		}
		if tagKey != "" {
			f.name = tagKey
		}
		if _, ok = flags[requiredFlag]; ok {
			f.required = true
		}
	}

	for _, key := range helpKeys {
		if tagKey, ok := sf.Tag.Lookup(key); ok {
			help, _ := reflectx.ParseTagKeyAndFlags(tagKey)
			f.helps[key] = help
		}
	}

	f.name = stringsx.LowerDashJoint(f.name)
	f.envkey = stringsx.UpperSnakeCase(f.name)

	return f
}

type Flag struct {
	name     string
	envkey   string
	lang     Lang
	helps    map[string]string
	required bool
	defaults any
	value    reflect.Value
}

func (f *Flag) Name() string {
	return f.name
}

func (f *Flag) LangHelp(lang Lang) string {
	key := strings.ToLower("help." + string(lang))
	keys := append([]string{key}, helpKeys...)

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

func (f *Flag) Help() string {
	return f.LangHelp(f.lang)
}

func (f *Flag) Env(prefix string) (string, string) {
	key := f.envkey
	if prefix != "" {
		key = strings.ToUpper(prefix + "__" + f.envkey)
	}
	return key, fmt.Sprint(f.defaults)
}

func (f *Flag) Register(cmd *cobra.Command) error {
	switch defaults := f.defaults.(type) {
	case bool:
		cmd.Flags().BoolVarP(f.value.Addr().Interface().(*bool), f.name, "", defaults, f.Help())
	case string:
		cmd.Flags().StringVarP(f.value.Addr().Interface().(*string), f.name, "", defaults, f.Help())
	case int:
		cmd.Flags().IntVarP(f.value.Addr().Interface().(*int), f.name, "", defaults, f.Help())
	case int8:
		cmd.Flags().Int8VarP(f.value.Addr().Interface().(*int8), f.name, "", defaults, f.Help())
	case int16:
		cmd.Flags().Int16VarP(f.value.Addr().Interface().(*int16), f.name, "", defaults, f.Help())
	case int32:
		cmd.Flags().Int32VarP(f.value.Addr().Interface().(*int32), f.name, "", defaults, f.Help())
	case int64:
		cmd.Flags().Int64VarP(f.value.Addr().Interface().(*int64), f.name, "", defaults, f.Help())
	case uint:
		cmd.Flags().UintVarP(f.value.Addr().Interface().(*uint), f.name, "", defaults, f.Help())
	case uint8:
		cmd.Flags().Uint8VarP(f.value.Addr().Interface().(*uint8), f.name, "", defaults, f.Help())
	case uint16:
		cmd.Flags().Uint16VarP(f.value.Addr().Interface().(*uint16), f.name, "", defaults, f.Help())
	case uint32:
		cmd.Flags().Uint32VarP(f.value.Addr().Interface().(*uint32), f.name, "", defaults, f.Help())
	case uint64:
		cmd.Flags().Uint64VarP(f.value.Addr().Interface().(*uint64), f.name, "", defaults, f.Help())
	case float32:
		cmd.Flags().Float32VarP(f.value.Addr().Interface().(*float32), f.name, "", defaults, f.Help())
	case float64:
		cmd.Flags().Float64VarP(f.value.Addr().Interface().(*float64), f.name, "", defaults, f.Help())
	case []int64:
		cmd.Flags().Int64SliceVarP(f.value.Addr().Interface().(*[]int64), f.name, "", defaults, f.Help())
	case []float64:
		cmd.Flags().Float64SliceVarP(f.value.Addr().Interface().(*[]float64), f.name, "", defaults, f.Help())
	case []string:
		cmd.Flags().StringSliceVarP(f.value.Addr().Interface().(*[]string), f.name, "", defaults, f.Help())
	default:
		return errors.Errorf("unsupport flag value type: `%s`", f.value.Type())
	}
	if f.required {
		return cmd.MarkFlagRequired(f.name)
	}
	return nil
}
