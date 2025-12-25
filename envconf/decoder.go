package envconf

import (
	"reflect"
	"strings"

	"github.com/xoctopus/x/codex"
	"github.com/xoctopus/x/reflectx"
	"github.com/xoctopus/x/textx"
)

func NewDecoder(g *Group) *Decoder {
	return &Decoder{g: g}
}

type Decoder struct {
	g *Group
}

func (d *Decoder) Decode(v any) error {
	pw := NewPathWalker()

	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}
	if !rv.IsValid() {
		return codex.Errorf(E_DEC__INVALID_VALUE, "at %s", pw.String())
	}

	return d.decode(pw, rv)
}

func (d *Decoder) decode(pw *PathWalker, rv reflect.Value) error {
	var (
		rt   = rv.Type()
		kind = rv.Kind()
	)

	if kind == reflect.Pointer {
		if rv.IsNil() {
			rv.Set(reflectx.New(rt))
		}
		return d.decode(pw, rv.Elem())
	}

	if !rv.CanSet() {
		return codex.Errorf(E_DEC__INVALID_VALUE_CANNOT_SET, "at %s", pw.String())
	}

	if dft, ok := rv.Addr().Interface().(interface{ SetDefault() }); ok {
		dft.SetDefault()
	}

	if reflect.PointerTo(rt).Implements(TextUnmarshallerT) {
		/*
			If I need this value implements TextUnmarshaler directly, but not
			through an embedded field which is a TextUnmarshaler implementation.
			But I have no clue to check this case.😂.
			So for avoiding this ambiguity, pls don't embed an TextMarshaler in
			a struct directly as an env configuration
		*/
		if v := d.g.Get(pw.String()); v != nil {
			return codex.Wrapf(E_DEC__FAILED_UNMARSHAL, textx.Unmarshal([]byte(v.val), rv), "at %s", pw.String())
		}
		return nil
	}

	if kind == reflect.Func || kind == reflect.Interface || kind == reflect.Chan {
		return nil // skip
	}

	switch kind {
	case reflect.Map:
		krt := rt.Key()
		vrt := rt.Elem()
		if k := krt.Kind(); !(reflectx.IsInteger(k) || k == reflect.String) {
			return codex.Errorf(E_DEC__INVALID_MAP_KEY_TYPE, "at %s[%s]", pw.String(), krt)
		}

		if rv.IsNil() {
			rv.Set(reflect.MakeMap(rt))
		}
		keys := d.g.MapEntries(pw.String())
		for _, key := range keys {
			krv := reflect.New(krt)
			if err := textx.Unmarshal([]byte(key), krv); err != nil {
				return codex.Wrapf(E_DEC__FAILED_UNMARSHAL, err, "at %s", pw.String())
			}
			// key = strings.ToUpper(key)
			pw.Enter(key)
			vrv := reflect.New(vrt)
			if err := d.decode(pw, vrv); err != nil {
				return err
			}
			pw.Leave()
			rv.SetMapIndex(krv.Elem(), vrv.Elem())
		}
		return nil
	case reflect.Slice, reflect.Array:
		length := d.g.SliceLength(pw.String())
		if length > 0 {
			if kind == reflect.Slice {
				rv.Set(reflect.MakeSlice(rt, length, length))
			} else {
				if rv.Len() < length {
					length = rv.Len()
				}
			}
		}
		for i := 0; i < length; i++ {
			pw.Enter(i)
			if d.g.Get(pw.String()) != nil {
				if err := d.decode(pw, rv.Index(i)); err != nil {
					return err
				}
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

			if flag := reflectx.ParseTag(frt.Tag).Get("env"); flag != nil {
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
			if err := d.decode(pw, rv.Field(i)); err != nil {
				return err
			}
			if !flatten {
				pw.Leave()
			}
		}
		return nil
	default:
		if v := d.g.Get(pw.String()); v != nil {
			return codex.Wrapf(E_DEC__FAILED_UNMARSHAL, textx.Unmarshal([]byte(v.val), rv), "at %s", pw.String())
		}
		return nil
	}
}
