package cmdx

import (
	"encoding"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/xoctopus/x/docx"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/ptrx"
	"github.com/xoctopus/x/reflectx"
	"github.com/xoctopus/x/stringsx"
	"github.com/xoctopus/x/textx"

	"github.com/xoctopus/confx/pkg/envx"
)

var (
	// tTextMarshaler   = reflect.TypeFor[encoding.TextMarshaler]()
	// tTextUnmarshaler = reflect.TypeFor[encoding.TextUnmarshaler]()
	tTextArshaler = reflect.TypeFor[interface {
		encoding.TextUnmarshaler
		encoding.TextMarshaler
	}]()
)

const (
	// OPTION_REQUIRE marks flag as required
	OPTION_REQUIRE = "require"

	// OPTION_PERSIST marks flag persistent to all sub commands
	OPTION_PERSIST = "persist"

	// OPTION_DEFAULT marks flag default value when flag is not presented
	OPTION_DEFAULT = "default"

	// OPTION_NOOP_DEFAULT marks flag default value when flag is presented but no option
	// eg: cmd --flag
	OPTION_NOOP_DEFAULT = "noopdef"

	// OPTION_SHORTHAND marks flag shorthand
	OPTION_SHORTHAND = "short"
)

// parseFlags parses flag from rv
func parseFlags(rv reflect.Value, pw *envx.PathWalker, walked map[string]map[string]*Flag) []*Flag {
	if rv.Kind() == reflect.Pointer {
		rv = reflectx.IndirectNew(rv)
	}
	rt := rv.Type()

	must.BeTrueF(rv.IsValid(), "expect valid value")
	must.BeTrueF(rv.Kind() == reflect.Struct, "expect a struct value, but got %s", rv.Type())
	must.BeTrueF(rv.CanSet(), "expect value can set")

	flags := make([]*Flag, 0)
	doc := func(...string) ([]string, bool) { return []string{}, false }

	if d, ok := rv.Addr().Interface().(docx.Doc); ok {
		doc = d.DocOf
	}

	for i := 0; i < rv.NumField(); i++ {
		frt := rv.Type().Field(i)
		if !frt.IsExported() {
			continue
		}
		name := frt.Name
		flag := reflectx.ParseTag(frt.Tag).Get("cmd")

		if flag != nil {
			if flag.Name() == "-" {
				continue
			}
			if flag.Name() != "" {
				name = flag.Name()
			}
		}

		if (frt.Anonymous && reflectx.Deref(frt.Type).Kind() == reflect.Struct) &&
			!(frt.Type.Implements(tTextArshaler) || reflect.PointerTo(frt.Type).Implements(tTextArshaler)) {
			if flag != nil {
				pw.Enter(name)
			}
			flags = append(flags, parseFlags(rv.Field(i), pw, walked)...)
			if flag != nil {
				pw.Leave()
			}
			continue
		}

		pw.Enter(name)

		frv := rv.Field(i)
		f := &Flag{
			fname: frt.Name,
			name:  pw.String(),
			value: reflectx.IndirectNew(frv),
		}

		if desc, ok := doc(f.fname); ok {
			f.usage = strings.Join(desc, " ")
		}

		if flag != nil {
			for opt := range flag.Options() {
				switch opt.Key() {
				case OPTION_PERSIST:
					f.persistent = true
				case OPTION_REQUIRE:
					f.required = true
				case OPTION_DEFAULT:
					if v := opt.Unquoted(); len(v) > 0 {
						f.defaults = ptrx.Ptr(strings.ToLower(v))
					}
				case OPTION_SHORTHAND:
					if v := opt.Unquoted(); len(v) > 0 {
						must.BeTrueF(
							len(v) == 1,
							"shorthand options must be on letter: %s.%s",
							rt.Name(), f.fname, v,
						)
						f.short = v
					}
				}
			}
			if o := flag.Option(OPTION_NOOP_DEFAULT); o != nil {
				f.noOptDef = ptrx.Ptr(o.Unquoted())
			} else {
				f.noOptDef = f.defaults
			}
		}

		_, ok := walked["cmd"][f.name]
		must.BeTrueF(!ok, "flag name conflict %s.%s [%s]", rt.Name(), f.fname, f.name)
		walked["cmd"][f.name] = f
		_, ok = walked["short"][f.short]
		must.BeTrueF(len(f.short) == 0 || !ok, "flag shorthand conflict %s.%s [%s]", rt.Name(), f.fname, f.short)
		walked["short"][f.short] = f

		flags = append(flags, f)

		pw.Leave()
	}
	return flags
}

// Flag implements pflag.Value
//
// it can receive all basic types, such as `int`, `uint`. and types that
// implement encoding.Unmarshaller/Marshaller.
// it is also acceptable if they are elements of a one-dimensional slice.
type Flag struct {
	fname      string
	name       string
	short      string
	usage      string
	required   bool
	persistent bool
	defaults   *string
	noOptDef   *string
	value      reflect.Value
}

func (f *Flag) Name() string {
	return f.name
}

func (f *Flag) Short() string {
	return f.short
}

func (f *Flag) Env() string {
	return stringsx.UpperCamelCase(f.name)
}

func (f *Flag) IsRequired() bool {
	return f.required
}

func (f *Flag) IsPersistent() bool {
	return f.persistent
}

func (f *Flag) Defaults() string {
	if f.defaults != nil {
		return *f.defaults
	}
	return ""
}

func (f *Flag) NoOptDefaults() string {
	if f.noOptDef == nil {
		return ""
	}
	return *f.noOptDef
}

func (f *Flag) Register(cmd *cobra.Command, envPrefix string) {
	f.name = stringsx.LowerDashJoint(f.name)

	// set from env
	envKey := strings.ToUpper(envPrefix + "__" + f.Env())
	envVal := os.Getenv(envKey)
	must.NoErrorF(f.Set(envVal), "failed to set '%s' from env %s=%s", f.name, envKey, envVal)

	// set default
	if v := f.Defaults(); len(v) > 0 && f.value.IsValid() && f.value.IsZero() {
		must.NoErrorF(f.Set(v), "failed to set '%s' as default '%s'", f.name, f.Defaults())
	}

	flag := &pflag.Flag{
		Name:     f.name,
		Value:    f,
		DefValue: f.String(),
	}

	flag.Usage = f.usage
	flag.NoOptDefVal = f.NoOptDefaults()

	if len(f.short) == 1 {
		flag.Shorthand = f.Short()
	}

	if f.IsPersistent() {
		cmd.PersistentFlags().AddFlag(flag)
	} else {
		cmd.Flags().AddFlag(flag)
	}

	if f.IsRequired() {
		must.NoErrorF(
			cmd.MarkFlagRequired(f.Name()),
			"failed to mark flag as required: %s, got error: %v", f.name,
		)
	}
}

func (f *Flag) String() string {
	if f.value.Kind() == reflect.Slice {
		fields := make([]string, f.value.Len())
		for i := 0; i < f.value.Len(); i++ {
			raw, err := textx.Marshal(f.value.Index(i))
			must.NoErrorF(
				err,
				"failed to marshal %s at index %d, got error: %v",
				f.Type(), i, err,
			)
			fields[i] = string(raw)
		}
		return "[" + strings.Join(fields, " ") + "]"
	}

	raw, err := textx.Marshal(f.value)
	must.NoErrorF(err, "failed to marshal %s", f.Type())

	return string(raw)
}

func (f *Flag) Set(val string) error {
	if val == "" {
		return nil
	}
	if f.value.Kind() == reflect.Slice {
		fields := strings.Fields(val)
		for i, field := range fields {
			v := reflect.New(f.value.Type().Elem()).Elem()
			if err := textx.Unmarshal([]byte(field), v); err != nil {
				return fmt.Errorf(
					"failed to parse %s at index %d by %s, got error: %w",
					f.Type(), i, field, err,
				)
			}
			f.value.Set(reflect.Append(f.value, v))
		}
	} else {
		if err := textx.Unmarshal([]byte(val), f.value); err != nil {
			return fmt.Errorf(
				"failed to parse %s from %s, got error: %w",
				f.Type(), val, err,
			)
		}
	}

	return nil
}

func (f *Flag) Type() string {
	return f.value.Type().String()
}
