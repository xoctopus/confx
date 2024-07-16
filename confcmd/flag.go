package confcmd

import (
	"reflect"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/misc/stringsx"
	"github.com/xoctopus/x/ptrx"
	"github.com/xoctopus/x/reflectx"
	"github.com/xoctopus/x/textx"

	"github.com/xoctopus/confx/envconf"
)

func ParseFlags(v any) []*Flag {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflectx.Indirect(reflect.ValueOf(v))
	}
	must.BeTrueWrap(rv != reflectx.InvalidValue, "invalid input value")

	flags := parseFlags(rv, envconf.NewPathWalker())
	sort.Slice(flags, func(i, j int) bool {
		return flags[i].name < flags[j].name
	})
	return flags
}

func parseFlags(rv reflect.Value, pw *envconf.PathWalker) []*Flag {
	if rv.Kind() == reflect.Pointer {
		rv = reflectx.IndirectNew(rv)
	}

	must.BeTrueWrap(rv.IsValid(), "expect valid value")
	must.BeTrueWrap(rv.Kind() == reflect.Struct, "expect a struct value, but got %s", rv.Type())
	must.BeTrueWrap(rv.CanSet(), "expect value can set")

	flags := make([]*Flag, 0)

	for i := 0; i < rv.NumField(); i++ {
		frt := rv.Type().Field(i)
		if !frt.IsExported() {
			continue
		}
		name := stringsx.LowerDashJoint(frt.Name)
		tag, flag := reflectx.ParseTagValue(frt.Tag.Get(tagCmd))
		if tag == "-" {
			continue
		}
		if tag != "" {
			name = tag
		}
		_, inlined := flag[flagInline]
		if reflectx.Deref(frt.Type).Kind() == reflect.Struct && !inlined {
			embed := frt.Anonymous && len(frt.Tag) == 0
			if !embed {
				pw.Enter(name)
			}
			flags = append(flags, parseFlags(rv.Field(i), pw)...)
			if !embed {
				pw.Leave()
			}
		} else {
			flags = append(flags, newFlag(pw, frt, rv.Field(i)))
		}
	}
	return flags
}

func newFlag(pw *envconf.PathWalker, sf reflect.StructField, fv reflect.Value) *Flag {
	f := &Flag{}

	{
		key, flags := reflectx.ParseTagValue(sf.Tag.Get(tagCmd))
		if key == "" {
			key = stringsx.LowerDashJoint(sf.Name)
		}
		key = strings.ToLower(key)
		f.name = strings.TrimPrefix(pw.FlagKey()+"-"+key, "-")
		for flag := range flags {
			kv := strings.Split(flag, "=")
			switch kv[0] {
			case flagRequire:
				f.required = true
			case flagPersist:
				f.persistent = true
			case flagDefault:
				if len(kv) == 2 {
					f.noOptDefValue = ptrx.Ptr(kv[1])
				}
			case flagShort:
				if len(kv) == 2 {
					f.short = kv[1]
				}
			}
		}
	}

	{
		key, _ := reflectx.ParseTagValue(sf.Tag.Get(tagEnv))
		if key != "-" {
			if key == "" {
				key = stringsx.UpperSnakeCase(sf.Name)
			}
			key = strings.ToUpper(key)
			f.env = ptrx.Ptr(strings.TrimPrefix(pw.EnvKey()+"_"+key, "_"))
		}
	}
	f.help = sf.Tag.Get(tagHelp)
	f.i18n = sf.Tag.Get(tagI18n)
	f.value = reflectx.IndirectNew(fv)
	return f
}

type Flag struct {
	name          string
	short         string
	env           *string
	help          string
	i18n          string
	required      bool
	persistent    bool
	noOptDefValue *string
	value         reflect.Value
}

func (f *Flag) Name() string { return f.name }

func (f *Flag) Short() string { return f.short }

func (f *Flag) Env() string {
	if f.env == nil {
		return ""
	}
	return *f.env
}

func (f *Flag) Help() string { return f.help }

func (f *Flag) I18n() string { return f.i18n }

func (f *Flag) IsRequired() bool { return f.required }

func (f *Flag) IsPersistent() bool { return f.persistent }

func (f *Flag) NoOptionDefaultValue() string {
	if f.noOptDefValue == nil {
		return ""
	}
	return *f.noOptDefValue
}

func (f *Flag) Register(cmd *cobra.Command, envVar string, localize LocalizeHelper) {
	fv := newFlagValue(f.value)

	must.NoErrorWrap(fv.Set(envVar), "failed to set env var: %s %s", f.name, envVar)

	flag := &pflag.Flag{
		Name:     f.name,
		Usage:    f.help,
		Value:    fv,
		DefValue: fv.String(),
	}
	if localize != nil && f.i18n != "" {
		if usage := localize(f.i18n); usage != "" {
			flag.Usage = usage
		}
	}

	if f.noOptDefValue != nil {
		flag.NoOptDefVal = *f.noOptDefValue
	}
	if len(f.short) == 1 {
		flag.Shorthand = f.short
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
}

func newFlagValue(rv reflect.Value) pflag.Value {
	return &flagValue{rv: rv}
}

// flagValue refers the Flag.value, implements pflag.Value
//
// it can receive all basic types, such as `int`, `uint`. and types that
// implement encoding.Unmarshaller/Marshaller.
// it is also acceptable if they are elements of a one-dimensional slice.
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
	if fv.rv.Kind() == reflect.Slice {
		fields := strings.Fields(val)
		for i, field := range fields {
			v := reflect.New(fv.rv.Type().Elem()).Elem()
			if err := textx.UnmarshalText([]byte(field), v); err != nil {
				return errors.Wrapf(
					err, "failed to parse %s at index %d by %s",
					fv.Type(), i, val,
				)
			}
			fv.rv.Set(reflect.Append(fv.rv, v))
		}
	} else {
		return errors.Wrapf(
			textx.UnmarshalText([]byte(val), fv.rv),
			"failed to parse %s from %s", fv.Type(), val,
		)
	}

	return nil
}

func (fv *flagValue) Type() string {
	return fv.rv.Type().String()
}
