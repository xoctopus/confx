package envconf

import (
	"reflect"

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

	return d.decode(pw, rv)
}

func (d *Decoder) decode(pw *PathWalker, rv reflect.Value) error {
	if !rv.IsValid() {
		return NewInvalidValueErr()
	}

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
		return NewCannotSetErr(rt)
	}

	if setter, ok := rv.Addr().Interface().(interface {
		SetDefault()
	}); ok {
		setter.SetDefault()
	}

	if rt.Implements(rtTextUnmarshaller) || reflect.PointerTo(rt).Implements(rtTextMarshaller) {
		if v := d.g.Get(pw.String()); v != nil {
			return textx.UnmarshalText([]byte(v.Value), rv)
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
		if k := krt.Kind(); !(k >= reflect.Int && k <= reflect.Uint64 || k == reflect.String) {
			return NewErrUnmarshalUnexpectMapKeyType(krt)
		}

		if rv.IsNil() {
			rv.Set(reflect.MakeMap(rt))
		}
		keys := d.g.MapEntries(pw.String())
		for _, key := range keys {
			krv := reflect.New(krt)
			if err := textx.UnmarshalText([]byte(key), krv); err != nil {
				return err
			}
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
		if kind == reflect.Slice && rv.IsNil() {
			rv.Set(reflect.MakeSlice(rt, length, length))
		}
		for i := 0; i < rv.Len(); i++ {
			pw.Enter(i)
			if err := d.decode(pw, rv.Index(i)); err != nil {
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
			flag := map[string]struct{}{}
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
			if err := d.decode(pw, rv.Field(i)); err != nil {
				return err
			}
			if !inline {
				pw.Leave()
			}
		}
		return nil
	default:
		if v := d.g.Get(pw.String()); v != nil {
			return textx.UnmarshalText([]byte(v.Value), rv)
		}
		return nil
	}
}
