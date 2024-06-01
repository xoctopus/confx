package confcmd

import (
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	return f.value.Interface()
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

func (f *Flag) Register(cmd *cobra.Command, lang LangType, envVar string) error {
	fv := NewFlagValue(f.value)

	if envVar != "" {
		if err := fv.Set(envVar); err != nil {
			return err
		}
	}

	cmd.Flags().AddFlag(&pflag.Flag{
		Name:     f.name,
		Usage:    f.Help(lang),
		Value:    fv,
		DefValue: fv.String(),
	})

	if f.required {
		return cmd.MarkFlagRequired(f.name)
	}

	return nil
}
