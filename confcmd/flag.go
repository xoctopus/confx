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
		parts := strings.Split(v, ",")
		if parts[0] == "-" {
			return nil
		}
		if parts[0] != "" {
			f.name = strings.TrimPrefix(prefix+"."+parts[0], ".")
		}
		for _, part := range parts[1:] {
			part = strings.TrimSpace(part)
			if part == FlagRequired || part == FlagRequiredShort {
				f.required = true
				continue
			}
			if part == FlagPersistent || part == FlagPersistentShort {
				f.persistent = true
				continue
			}
			kv := strings.Split(part, "=")
			switch strings.ToLower(kv[0]) {
			case FlagCanNoOptShort, FlagCanNoOpt:
				if len(kv) == 2 {
					f.noOptionDef = ptrx.Ptr(kv[1])
				}
			case FlagShorthandShort, FlagShorthand:
				if len(kv) == 2 {
					f.shorthand = kv[1]
				}
			}
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
	// shorthand
	shorthand string
	// env key can injector from env var, eg: SOME_FLAG
	env *string
	// helps multi-language help info
	helps map[string]string
	// required if this flag is required
	required bool
	// persistent if this flag is persistent
	persistent bool
	// noOptionDef means the default when a flag has no option.
	// like `uname --kernel-version` the flag has no option default to enable
	// print kernel version
	noOptionDef *string
	// value filed value
	value reflect.Value
}

func (f *Flag) Name() string {
	return f.name
}

func (f *Flag) Shorthand() string {
	return f.shorthand
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

func (f *Flag) IsPersistent() bool {
	return f.persistent
}

func (f *Flag) DefaultValue() any {
	return f.value.Interface()
}

func (f *Flag) NoOptionDefaultValue() string {
	if f.noOptionDef == nil {
		return ""
	}
	return *f.noOptionDef
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

	flag := &pflag.Flag{
		Name:     f.name,
		Usage:    f.Help(lang),
		Value:    fv,
		DefValue: fv.String(),
	}

	if f.noOptionDef != nil {
		flag.NoOptDefVal = *f.noOptionDef
	}
	if len(f.shorthand) == 1 {
		flag.Shorthand = f.shorthand
	}

	if f.persistent {
		cmd.PersistentFlags().AddFlag(flag)
	} else {
		cmd.Flags().AddFlag(flag)
	}

	if f.required {
		must.NoErrorWrap(
			cmd.MarkFlagRequired(f.name),
			"failed to mark flag as required: %s", f.name,
		)
	}

	return nil
}
