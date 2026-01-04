package envx_test

import (
	"fmt"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/envx"
)

func TestEncoder_Encode(t *testing.T) {
	t.Run("Invalid", func(t *testing.T) {
		t.Run("InvalidGroupName", func(t *testing.T) {
			ExpectPanic[string](t, func() { envx.NewGroup("x.x") })

		})
		grp := envx.NewGroup("TEST")
		enc := envx.NewEncoder(grp)

		err := enc.Encode(nil)
		Expect(t, err, Succeed())
		Expect(t, grp.Len(), Equal(0))
	})

	t.Run("IgnoredKinds", func(t *testing.T) {
		grp := envx.NewGroup("TEST")
		enc := envx.NewEncoder(grp)

		err := enc.Encode(&struct {
			Func func()
			Chan chan struct{}
			Any  any
		}{})
		Expect(t, err, Succeed())
	})

	t.Run("Map", func(t *testing.T) {
		t.Run("UnexpectMapKeyType", func(t *testing.T) {
			grp := envx.NewGroup("TEST")
			enc := envx.NewEncoder(grp)
			// pos := envconf.NewPathWalker()

			type Key struct{}
			err := enc.Encode(map[Key]string{})
			Expect(t, err, IsCodeError(envx.CODE__ENC_INVALID_MAP_KEY_TYPE))
		})
		t.Run("UnexpectMapKeyValue", func(t *testing.T) {
			grp := envx.NewGroup("TEST")
			enc := envx.NewEncoder(grp)
			// pos := envconf.NewPathWalker()
			expect := envx.CODE__ENC_INVALID_MAP_KEY_VALUE
			t.Run("InvalidInteger", func(t *testing.T) {
				target := enc.Encode(map[int]string{-1: "any"})
				Expect(t, target, IsCodeError(expect))
			})
			t.Run("InvalidString", func(t *testing.T) {
				target := enc.Encode(map[string]string{"": "any"})
				Expect(t, target, IsCodeError(expect))

				target = enc.Encode(map[string]string{"a b c": "any"})
				Expect(t, target, IsCodeError(expect))
			})
		})
		t.Run("FailedToMarshal", func(t *testing.T) {
			grp := envx.NewGroup("TEST")
			enc := envx.NewEncoder(grp)
			pos := envx.NewPathWalker()

			pos.Enter(0)
			expect := envx.CODE__ENC_FAILED_MARSHAL
			target := enc.Encode(map[int]MustFailedArshaler{0: {}})
			Expect(t, target, IsCodeError(expect))
		})

		t.Run("Success", func(t *testing.T) {
			t.Run("StringKey", func(t *testing.T) {
				grp := envx.NewGroup("TEST")
				enc := envx.NewEncoder(grp)
				val := map[string]int{"1": 1, "2": 2}

				err := enc.Encode(val)
				Expect(t, err, Succeed())
				Expect(t, grp.Len(), Equal(2))
				Expect(t, grp.Len(), Equal(len(val)))
				for k := range val {
					v := grp.Get(k)
					Expect(t, fmt.Sprint(val[k]), Equal(v.Value()))
					Expect(t, fmt.Sprint(k), Equal(v.Key()))
				}
			})
			t.Run("IntegerKey", func(t *testing.T) {
				grp := envx.NewGroup("TEST")
				enc := envx.NewEncoder(grp)
				val := map[int]string{1: "1", 2: "2"}

				err := enc.Encode(val)
				Expect(t, err, Succeed())
				Expect(t, grp.Len(), Equal(2))
				Expect(t, grp.Len(), Equal(len(val)))
				for k := range val {
					v := grp.Get(fmt.Sprint(k))
					Expect(t, fmt.Sprint(val[k]), Equal(v.Value()))
					Expect(t, fmt.Sprint(k), Equal(v.Key()))
				}
			})
			t.Run("MapMap", func(t *testing.T) {
				grp := envx.NewGroup("TEST")
				enc := envx.NewEncoder(grp)
				val := map[int]map[string]int{
					1: {"1": 1},
					2: {"2": 2},
				}

				err := enc.Encode(val)
				Expect(t, err, Succeed())
				Expect(t, grp.Len(), Equal(2))
				Expect(t, grp.Len(), Equal(len(val)))
				Expect(t, grp.Get("1_1").Value(), Equal("1"))
				Expect(t, grp.Get("2_2").Value(), Equal("2"))
			})
			t.Run("NilMap", func(t *testing.T) {
				grp := envx.NewGroup("TEST")
				enc := envx.NewEncoder(grp)
				err := enc.Encode(map[int]int(nil))
				Expect(t, err, Succeed())
				Expect(t, grp.Len(), Equal(0))
			})
		})
	})

	t.Run("Slice", func(t *testing.T) {
		t.Run("Failed", func(t *testing.T) {
			grp := envx.NewGroup("TEST")
			enc := envx.NewEncoder(grp)

			pos := envx.NewPathWalker()
			pos.Enter(0)

			expect := envx.CODE__ENC_FAILED_MARSHAL
			target := enc.Encode([]MustFailedArshaler{{}})
			Expect(t, target, IsCodeError(expect))
		})
		grp := envx.NewGroup("TEST")
		enc := envx.NewEncoder(grp)

		err := enc.Encode([]int{1, 2, 3})
		Expect(t, err, Succeed())
	})

	t.Run("Struct", func(t *testing.T) {
		t.Run("FailedEncodeField", func(t *testing.T) {
			grp := envx.NewGroup("TEST")
			enc := envx.NewEncoder(grp)
			pos := envx.NewPathWalker()
			pos.Enter("X")

			expect := envx.CODE__ENC_FAILED_MARSHAL
			target := enc.Encode(struct {
				X MustFailedArshaler
			}{
				X: MustFailedArshaler{},
			})
			Expect(t, target, IsCodeError(expect))
		})
		t.Run("DuplicatedGroupKey", func(t *testing.T) {
			grp := envx.NewGroup("TEST")
			grp.Add(envx.NewVar("X", "100"))

			enc := envx.NewEncoder(grp)
			pos := envx.NewPathWalker()
			pos.Enter("X")

			target := enc.Encode(struct {
				NIL *int
				X   int `env:"x"`
				Y   int `env:"x,optional"`
			}{
				X: 100,
				Y: 101,
			})
			Expect(t, target, Succeed())
			Expect(t, grp.Len(), Equal(1))
			v := grp.Get("X")
			Expect(t, v, NotBeNil[*envx.Var]())
			Expect(t, v.Key(), Equal("X"))
			Expect(t, v.Value(), Equal("101"))
			Expect(t, v.Optional(), BeTrue())
		})
	})
}
