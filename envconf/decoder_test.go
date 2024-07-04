package envconf_test

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/xoctopus/datatypex"
	"github.com/xoctopus/x/ptrx"
	"github.com/xoctopus/x/textx"

	"github.com/xoctopus/confx/envconf"
)

func ExampleDecoder_Decode() {
	grp := envconf.NewGroup("TEST")

	grp.Add(envconf.NewVar("MapStringString_Key", "Value"))
	grp.Add(envconf.NewVar("MapStringInt_Key", "100"))
	grp.Add(envconf.NewVar("MapStringInt64_Key", "101"))
	grp.Add(envconf.NewVar("MapStringPassword_Key", "password"))
	grp.Add(envconf.NewVar("Slice_0_A", "2"))
	grp.Add(envconf.NewVar("Slice_0_B", "string"))
	grp.Add(envconf.NewVar("Array_0_A", "string"))
	grp.Add(envconf.NewVar("Array_0_B", "2"))

	dec := envconf.NewDecoder(grp)
	enc := envconf.NewEncoder(grp)

	v := &struct {
		MapStringString   map[string]string
		MapStringInt      map[string]int
		MapStringInt64    map[string]int64
		MapStringPassword map[string]datatypex.Password
		Slice             []struct {
			A int
			B string
		}
		Array [1]struct {
			B int
			A string
		}
	}{}

	if err := dec.Decode(v); err != nil {
		return
	}
	if err := enc.Encode(v); err != nil {
		return
	}

	demoVar := envconf.NewVar("VarName", "")
	fmt.Println(demoVar.GroupName(grp.Name))
	fmt.Println(string(grp.Bytes()))
	fmt.Println(string(grp.MaskBytes()))

	// output:
	// TEST__VarName
	// TEST__Array_0_A=string
	// TEST__Array_0_B=2
	// TEST__MapStringInt64_Key=101
	// TEST__MapStringInt_Key=100
	// TEST__MapStringPassword_Key=password
	// TEST__MapStringString_Key=Value
	// TEST__Slice_0_A=2
	// TEST__Slice_0_B=string
	//
	// TEST__Array_0_A=string
	// TEST__Array_0_B=2
	// TEST__MapStringInt64_Key=101
	// TEST__MapStringInt_Key=100
	// TEST__MapStringPassword_Key=--------
	// TEST__MapStringString_Key=Value
	// TEST__Slice_0_A=2
	// TEST__Slice_0_B=string
}

type DefaultSetter struct {
	Int int
}

func (v *DefaultSetter) SetDefault() {
	v.Int = 101
}

var (
	ErrMustUnmarshalFailed = errors.New("must unmarshal failed")
	ErrMustMarshalFailed   = errors.New("must marshal failed")
)

type MustFailed struct{}

func (failed *MustFailed) UnmarshalText([]byte) error {
	return ErrMustUnmarshalFailed
}

func (failed MustFailed) MarshalText() ([]byte, error) {
	return nil, ErrMustMarshalFailed
}

type Inline struct {
	String string
	Int    int
}

func TestDecoder_Decode(t *testing.T) {
	grp := envconf.NewGroup("TEST")
	dec := envconf.NewDecoder(grp)

	t.Run("InvalidValue", func(t *testing.T) {
		err := dec.Decode(nil)
		NewWithT(t).Expect(err).NotTo(BeNil())
		NewWithT(t).Expect(errors.Is(err, envconf.NewInvalidValueErr())).To(BeTrue())
	})

	t.Run("NeedSetDefault", func(t *testing.T) {
		grp.Reset()

		grp.Add(envconf.NewVar("Int", "100"))
		val := &DefaultSetter{}
		err := dec.Decode(val)
		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(val.Int).To(Equal(100))

		grp.Del("Int")

		err = dec.Decode(val)
		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(val.Int).To(Equal(101))
	})

	t.Run("CannotSet", func(t *testing.T) {
		err := dec.Decode(1)
		NewWithT(t).Expect(err).NotTo(BeNil())
		NewWithT(t).Expect(errors.Is(err, envconf.NewCannotSetErr(reflect.TypeOf(1)))).To(BeTrue())
	})

	t.Run("SkippedKinds", func(t *testing.T) {
		grp.Reset()

		val1 := &struct {
			Chan chan struct{}
		}{}

		grp.Add(envconf.NewVar("Chan", ""))
		NewWithT(t).Expect(dec.Decode(val1)).To(BeNil())

		val2 := &struct {
			Func func()
		}{}

		grp.Add(envconf.NewVar("Func", ""))
		NewWithT(t).Expect(dec.Decode(val2)).To(BeNil())

		val3 := &struct {
			Any any
		}{}

		grp.Add(envconf.NewVar("Any", ""))
		NewWithT(t).Expect(dec.Decode(val3)).To(BeNil())
	})

	t.Run("Map", func(t *testing.T) {
		t.Run("UnexpectedMapKeyType", func(t *testing.T) {
			grp.Reset()

			type Key struct {
				key string
			}

			val := &struct {
				Map map[Key]any
			}{}

			err := dec.Decode(val)
			NewWithT(t).Expect(err).NotTo(BeNil())
			NewWithT(t).Expect(errors.Is(err, envconf.NewErrUnmarshalUnexpectMapKeyType(reflect.TypeOf(Key{})))).To(BeTrue())
		})

		t.Run("FailedToUnmarshalMapKey", func(t *testing.T) {
			grp.Reset()
			grp.Add(envconf.NewVar("Map_any_non_numeric", ""))

			val := &struct {
				Map map[int]any
			}{}

			err := dec.Decode(val)
			NewWithT(t).Expect(err).NotTo(BeNil())
			NewWithT(t).Expect(errors.As(err, ptrx.Ptr(&textx.ErrUnmarshalFailed{}))).To(BeTrue())
		})

		t.Run("FailedToUnmarshalMapValue", func(t *testing.T) {
			grp.Reset()
			grp.Add(envconf.NewVar("Map_100", "any_non_numeric"))

			val := &struct {
				Map map[int]int
			}{}

			err := dec.Decode(val)
			NewWithT(t).Expect(err).NotTo(BeNil())
			NewWithT(t).Expect(errors.As(err, ptrx.Ptr(&textx.ErrUnmarshalFailed{}))).To(BeTrue())
		})

		t.Run("Success", func(t *testing.T) {
			grp.Reset()
			grp.Add(envconf.NewVar("Map_key1", "val1"))
			grp.Add(envconf.NewVar("Map_key2", "val2"))
			grp.Add(envconf.NewVar("Map_key3", "val3"))

			val := &struct {
				Map map[string]string
			}{}

			err := dec.Decode(val)
			NewWithT(t).Expect(err).To(BeNil())
			NewWithT(t).Expect(val.Map).To(Equal(map[string]string{
				"key1": "val1",
				"key2": "val2",
				"key3": "val3",
			}))
		})
	})

	t.Run("Slice", func(t *testing.T) {
		t.Run("FailedToDecodeSlice", func(t *testing.T) {
			grp.Reset()
			grp.Add(envconf.NewVar("Slice_0", "any_non_numeric"))

			val := &struct {
				Slice []int
			}{}

			err := dec.Decode(val)
			NewWithT(t).Expect(err).NotTo(BeNil())
			NewWithT(t).Expect(errors.As(err, ptrx.Ptr(&textx.ErrUnmarshalFailed{}))).To(BeTrue())
		})
		t.Run("FailedToDecodeArray", func(t *testing.T) {
			grp.Reset()
			grp.Add(envconf.NewVar("Array_0", "any_non_numeric"))

			val := &struct {
				Array [1]int
			}{}

			err := dec.Decode(val)
			NewWithT(t).Expect(err).NotTo(BeNil())
			NewWithT(t).Expect(errors.As(err, ptrx.Ptr(&textx.ErrUnmarshalFailed{}))).To(BeTrue())
		})

		t.Run("Success", func(t *testing.T) {
			grp.Reset()
			grp.Add(envconf.NewVar("Slice_0", "0"))
			// skip index 1 grp.Add(envconf.NewVar("Slice_1", "1"))
			grp.Add(envconf.NewVar("Slice_2", "2"))
			grp.Add(envconf.NewVar("Array_0", "0"))
			// skip index 1 grp.Add(envconf.NewVar("Array_1", "1"))
			grp.Add(envconf.NewVar("Array_2", "2"))

			val := &struct {
				Slice []int
				Array [3]string
			}{}

			err := dec.Decode(val)
			NewWithT(t).Expect(err).To(BeNil())
			NewWithT(t).Expect(val.Slice).To(Equal([]int{0, 0, 2}))
			NewWithT(t).Expect(val.Array).To(Equal([3]string{"0", "", "2"}))
		})
	})

	t.Run("Struct", func(t *testing.T) {
		grp.Reset()
		grp.Add(envconf.NewVar("Endpoint", "https://host:1111/abcd"))
		grp.Add(envconf.NewVar("SomeEndpoint", "https://host:2222/abcd"))
		grp.Add(envconf.NewVar("private", "1"))
		grp.Add(envconf.NewVar("Inline_String", "inline_string_1"))
		grp.Add(envconf.NewVar("Inline_Int", "1"))
		grp.Add(envconf.NewVar("String", "inline_string_2"))
		grp.Add(envconf.NewVar("Int", "2"))
		grp.Add(envconf.NewVar("Skip", "100"))
		grp.Add(envconf.NewVar("IntPtr", "100"))
		grp.Add(envconf.NewVar("Field_String", "field_string"))
		grp.Add(envconf.NewVar("Field_Int", "3"))
		grp.Add(envconf.NewVar("Failed", "any"))

		val := &struct {
			private  int
			Endpoint datatypex.Endpoint `env:"SomeEndpoint"`
			Field    Inline
			Skip     int `env:"-"`
			IntPtr   *int
			Inline
			Timestamp datatypex.Timestamp
			Failed    MustFailed
		}{
			private: 19820, // to check if this value is modified
			Skip:    19820, // to check if this value is modified
		}

		err := dec.Decode(val)
		NewWithT(t).Expect(err).NotTo(BeNil())
		NewWithT(t).Expect(errors.As(err, ptrx.Ptr(&textx.ErrUnmarshalFailed{}))).To(BeTrue())
		NewWithT(t).Expect(err.Error()).To(ContainSubstring(ErrMustUnmarshalFailed.Error()))

		NewWithT(t).Expect(val.Skip).To(Equal(19820))
		NewWithT(t).Expect(val.private).To(Equal(19820))
		NewWithT(t).Expect(val.Endpoint.Port).To(Equal(uint16(2222)))
		NewWithT(t).Expect(val.String).To(Equal("inline_string_2"))
		NewWithT(t).Expect(val.Int).To(Equal(2))
		NewWithT(t).Expect(val.Field.String).To(Equal("field_string"))
		NewWithT(t).Expect(val.Field.Int).To(Equal(3))
		NewWithT(t).Expect(*val.IntPtr).To(Equal(100))
	})
}
