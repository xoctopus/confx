package envconf

import (
	"reflect"
	"strings"

	"github.com/xoctopus/x/codex"
	"github.com/xoctopus/x/reflectx"
	"github.com/xoctopus/x/textx"
)

func NewEncoder(g *Group) *Encoder {
	return &Encoder{g: g, flags: make(map[string]*reflectx.Flag)}
}

type Encoder struct {
	g     *Group
	flags map[string]*reflectx.Flag
}

func (d *Encoder) Encode(v any) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	return d.encode(NewPathWalker(), rv)
}

func (d *Encoder) set(pw *PathWalker, rv reflect.Value) error {
	v := &Var{key: pw.String()}

	if flag, exists := d.flags[v.key]; exists && flag != nil {
		for opt := range flag.Options() {
			switch opt.Key() {
			case "optional":
				v.optional = true
			}
		}
	}

	text, err := textx.Marshal(rv)
	if err != nil {
		return codex.Wrapf(E_ENC__FAILED_MARSHAL, err, "at %s", pw.String())
	}
	v.val = string(text)
	v.mask = v.val

	if masker, ok := rv.Interface().(interface{ SecurityString() string }); ok {
		v.mask = masker.SecurityString()
	}

	// overwritten is allowed
	// if d.g.Add(v) {
	// 	return NewErrorf(pw, E_ENC__DUPLICATE_GROUP_KEY, "`%s`", d.g.Key(v.key))
	// }
	d.g.Add(v)
	return nil
}

func (d *Encoder) encode(pw *PathWalker, rv reflect.Value) error {
	if !rv.IsValid() {
		return nil
	}

	var (
		rt   = rv.Type()
		kind = rv.Kind()
	)

	if kind == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		return d.encode(pw, rv.Elem())
	}

	if rt.Implements(TextMarshallerT) || reflect.PointerTo(rt).Implements(TextMarshallerT) {
		return d.set(pw, rv)
	}

	if kind == reflect.Func || kind == reflect.Chan || kind == reflect.Interface {
		return nil // skip
	}

	switch kind {
	case reflect.Map:
		if rv.IsNil() {
			return nil
		}
		if k := rt.Key().Kind(); k != reflect.String && !reflectx.IsInteger(k) {
			return codex.Errorf(E_ENC__INVALID_MAP_KEY_TYPE, "at %s[%s]", pw, k)
		}
		keys := rv.MapKeys()
		for i := range keys {
			key := keys[i]
			keyv := key.Interface()
			switch {
			case key.Kind() == reflect.String:
				if x := key.String(); len(x) == 0 || !alphabet(x) {
					return codex.Errorf(E_ENC__INVALID_MAP_KEY_VALUE, "at %s[%s]", pw, keyv)
				}
			case key.CanInt():
				if key.Int() < 0 {
					return codex.Errorf(E_ENC__INVALID_MAP_KEY_VALUE, "at %s[%s]", pw, keyv)
				}
			}
			// keyv = strings.ToUpper(fmt.Sprint(keyv))
			pw.Enter(keyv)
			vrv := rv.MapIndex(keys[i])
			if err := d.encode(pw, vrv); err != nil {
				return err
			}
			pw.Leave()
		}
		return nil
	case reflect.Array, reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			pw.Enter(i)
			if err := d.encode(pw, rv.Index(i)); err != nil {
				return err
			}
			pw.Leave()
		}
		return nil
	case reflect.Struct:
		for i := 0; i < rv.NumField(); i++ {
			frt := rt.Field(i)
			if !frt.IsExported() {
				continue
			}
			name := frt.Name
			flatten := frt.Anonymous && reflectx.Deref(frt.Type).Kind() == reflect.Struct

			flag := reflectx.ParseTag(frt.Tag).Get("env")
			if flag != nil {
				flatten = false
				if flag.Name() == "-" {
					continue
				}
				if flag.Name() != "" {
					name = flag.Name()
					name = strings.ToUpper(name)
				}
			}

			if !flatten {
				pw.Enter(name)
			}

			if key := pw.String(); len(key) > 0 && !flatten && flag != nil {
				d.flags[key] = flag
			}

			if err := d.encode(pw, rv.Field(i)); err != nil {
				return err
			}
			if !flatten {
				pw.Leave()
			}
		}
		return nil
	default:
		return d.set(pw, rv)
	}
}

func alphabet(v string) bool {
	for _, c := range v {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}
