package envconf

import (
	"reflect"

	"github.com/xoctopus/x/reflectx"
	"github.com/xoctopus/x/textx"
)

func NewEncoder(g *Group) *Encoder {
	return &Encoder{g: g, flags: make(map[string]map[string]struct{})}
}

type Encoder struct {
	g     *Group
	flags map[string]map[string]struct{}
}

func (d *Encoder) Encode(v any) error {
	pw := NewPathWalker()

	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	return d.encode(pw, rv)
}

func (d *Encoder) set(pw *PathWalker, rv reflect.Value) error {
	v := &Var{Name: pw.String()}

	if flag := d.flags[v.Name]; len(flag) > 0 {
		v.ParseOption(flag)
	}

	if masker, ok := rv.Interface().(interface {
		SecurityString() string
	}); ok {
		v.Mask = masker.SecurityString()
	}
	text, err := textx.MarshalText(rv)
	if err != nil {
		return err
	}
	v.Value = string(text)

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

	if rt.Implements(rtTextMarshaller) ||
		rv.CanAddr() && rv.Addr().Type().Implements(rtTextMarshaller) {
		return d.set(pw, rv)
	}

	if kind == reflect.Func || kind == reflect.Chan || kind == reflect.Interface {
		return nil // skip
	}

	switch kind {
	case reflect.Map:
		krt := rt.Key()
		if k := krt.Kind(); !(k >= reflect.Int && k <= reflect.Uint64 || k == reflect.String) {
			return NewErrMarshalUnexpectMapKeyType(krt)
		}
		if rv.IsNil() {
			return nil
		}
		keys := rv.MapKeys()
		for i := range keys {
			key := keys[i].Interface()
			if _key, ok := key.(string); ok {
				if len(_key) == 0 || !alphabet(_key) {
					return ErrUnexpectMapKeyValue(_key)
				}
			}
			pw.Enter(key)
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
			flag := make(map[string]struct{})
			if tag, ok := frt.Tag.Lookup("env"); ok {
				tagName := ""
				tagName, flag = reflectx.ParseTagKeyAndFlags(tag)
				if tagName == "-" {
					continue
				}
				if tagName != "" {
					name = tagName
				}
			}
			inline := len(flag) == 0 && frt.Anonymous &&
				reflectx.Deref(frt.Type).Kind() == reflect.Struct
			if !inline {
				pw.Enter(name)
			}
			d.flags[pw.String()] = flag
			if err := d.encode(pw, rv.Field(i)); err != nil {
				return err
			}
			if !inline {
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
