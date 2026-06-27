package envx_test

import (
	"errors"
	"testing"
	"time"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/envx"
)

type DefaultSetter struct {
	Int int
}

func (v *DefaultSetter) SetDefault() {
	v.Int = 101
}

type MustFailedArshaler struct {
	V any
}

func (MustFailedArshaler) MarshalText() ([]byte, error) {
	return nil, errors.New("MUST FAILED MARSHALER")
}

func (*MustFailedArshaler) UnmarshalText([]byte) error {
	return errors.New("MUST FAILED UNMARSHALER")
}

func TestDecoder_Decode(t *testing.T) {
	t.Run("Invalid", func(t *testing.T) {
		grp := envx.NewGroup("TEST")
		Expect(t, grp.Name(), Equal("TEST"))
		Expect(t, grp.Values(), HaveLen[map[string]*envx.Var](0))

		dec := envx.NewDecoder(grp)

		t.Run("InvalidValue", func(t *testing.T) {
			err := dec.Decode(nil)
			Expect(t, err, IsCodeError(envx.CODE__DEC_INVALID_VALUE))
		})
		t.Run("CannotSet", func(t *testing.T) {
			err := dec.Decode(struct{}{})
			Expect(t, err, IsCodeError(envx.CODE__DEC_INVALID_VALUE_CANNOT_SET))
		})
	})

	t.Run("DefaultSetter", func(t *testing.T) {
		grp := envx.NewGroup("TEST")
		dec := envx.NewDecoder(grp)

		// no var for decoding, set as default
		val := &DefaultSetter{}
		Expect(t, dec.Decode(val), Succeed())
		Expect(t, val.Int, Equal(101))

		// overwritten by env var
		grp.Add(envx.NewVar("Int", "200"))
		Expect(t, dec.Decode(val), Succeed())
		Expect(t, val.Int, Equal(200))
	})

	t.Run("SkippedTypes", func(t *testing.T) {
		grp := envx.NewGroup("TEST")
		dec := envx.NewDecoder(grp)

		err := dec.Decode(&struct {
			Chan chan struct{}
			Func func()
			Any  any
		}{
			Chan: make(chan struct{}),
			Func: func() {},
			Any:  1,
		})
		Expect(t, err, Succeed())
	})

	t.Run("Map", func(t *testing.T) {
		t.Run("InvalidMapKeyType", func(t *testing.T) {
			grp := envx.NewGroup("TEST")
			dec := envx.NewDecoder(grp)
			pos := envx.NewPathWalker()

			type Key struct{}

			pos.Enter("Map")
			err := dec.Decode(&struct {
				Map map[Key]any
			}{})
			Expect(t, err, IsCodeError(envx.CODE__DEC_INVALID_MAP_KEY_TYPE))
		})
		t.Run("FailedToUnmarshalMapKey", func(t *testing.T) {
			grp := envx.NewGroup("TEST")
			dec := envx.NewDecoder(grp)
			pos := envx.NewPathWalker()

			grp.Add(envx.NewVar("Map_INVALIDINT", "any"))
			grp.Add(envx.NewVar("OTHER_VAR_KEY", "any"))
			pos.Enter("Map")

			err := dec.Decode(&struct {
				Map map[int]string
			}{})
			Expect(t, err, IsCodeError(envx.CODE__DEC_FAILED_UNMARSHAL))
		})
		t.Run("FailedToDecodeMapValue", func(t *testing.T) {
			grp := envx.NewGroup("TEST")
			dec := envx.NewDecoder(grp)
			pos := envx.NewPathWalker()

			pos.Enter("Map")
			pos.Enter("KEY1")
			grp.Add(envx.NewVar("Map_KEY1", "invalid integer string"))

			err := dec.Decode(&struct {
				Map map[string]int
			}{})
			Expect(t, err, IsCodeError(envx.CODE__DEC_FAILED_UNMARSHAL))
		})
		t.Run("Success", func(t *testing.T) {
			t.Run("RawMap", func(t *testing.T) {
				grp := envx.NewGroup("TEST")
				dec := envx.NewDecoder(grp)

				grp.Add(envx.NewVar("1", "1"))
				grp.Add(envx.NewVar("2", "2"))

				target := map[int]string{}
				expect := map[int]string{1: "1", 2: "2"}

				err := dec.Decode(&target)
				Expect(t, err, Succeed())
				Expect(t, expect, Equal(target))
			})
			t.Run("PrefixedMap", func(t *testing.T) {
				grp := envx.NewGroup("TEST")
				dec := envx.NewDecoder(grp)

				grp.Add(envx.NewVar("Map_1", "1"))
				grp.Add(envx.NewVar("Map_2", "2"))

				target := struct{ Map map[int]string }{}
				expect := struct{ Map map[int]string }{
					Map: map[int]string{1: "1", 2: "2"},
				}

				err := dec.Decode(&target)
				Expect(t, err, Succeed())
				Expect(t, expect, Equal(target))
			})
			t.Run("PrefixedAndSuffixedMap", func(t *testing.T) {
				grp := envx.NewGroup("TEST")
				dec := envx.NewDecoder(grp)

				grp.Add(envx.NewVar("Map_1_1", "1"))
				grp.Add(envx.NewVar("Map_2_2", "2"))

				target := struct{ Map map[int]map[string]int }{}
				expect := struct{ Map map[int]map[string]int }{
					Map: map[int]map[string]int{1: {"1": 1}, 2: {"2": 2}},
				}

				err := dec.Decode(&target)
				Expect(t, err, Succeed())
				Expect(t, expect, Equal(target))
			})
		})
	})

	t.Run("ArrayAndSlice", func(t *testing.T) {
		t.Run("FailedToDecodeElem", func(t *testing.T) {
			grp := envx.NewGroup("TEST")
			dec := envx.NewDecoder(grp)
			pos := envx.NewPathWalker()

			pos.Enter("0")
			grp.Add(envx.NewVar("0", "invalid integer string"))

			err := dec.Decode(&([]int{}))
			Expect(t, err, IsCodeError(envx.CODE__DEC_FAILED_UNMARSHAL))
		})
		t.Run("Success", func(t *testing.T) {
			t.Run("RawSlice", func(t *testing.T) {
				grp := envx.NewGroup("TEST")
				dec := envx.NewDecoder(grp)

				grp.Add(envx.NewVar("2", "2"))
				grp.Add(envx.NewVar("4", "4"))
				grp.Add(envx.NewVar("INVALID_INDEX", "any")) // skipped

				target := make([]int, 0)
				expect := []int{0, 0, 2, 0, 4}
				Expect(t, dec.Decode(&target), Succeed())
				Expect(t, expect, Equal(target))
			})
			t.Run("PrefixedArray", func(t *testing.T) {
				grp := envx.NewGroup("TEST")
				dec := envx.NewDecoder(grp)

				grp.Add(envx.NewVar("Array_2", "2"))
				grp.Add(envx.NewVar("Array_4", "4"))               // skipped: out of capacity
				grp.Add(envx.NewVar("Array_INVALID_INDEX", "any")) // skipped

				target := struct{ Array [3]int }{}
				expect := struct{ Array [3]int }{Array: [3]int{0, 0, 2}}
				Expect(t, dec.Decode(&target), Succeed())
				Expect(t, expect, Equal(target))
			})
			t.Run("PrefixedAndSuffixedArray", func(t *testing.T) {
				grp := envx.NewGroup("TEST")
				dec := envx.NewDecoder(grp)

				grp.Add(envx.NewVar("Array_1", "http://localhost:9999"))
				grp.Add(envx.NewVar("Array_4", "4"))               // skipped: out of capacity
				grp.Add(envx.NewVar("Array_INVALID_INDEX", "any")) // skipped

				target := struct{ Array [3]string }{}
				expect := struct{ Array [3]string }{Array: [3]string{"", "http://localhost:9999", ""}}
				Expect(t, dec.Decode(&target), Succeed())
				Expect(t, expect, Equal(target))
			})
		})
	})

	t.Run("Struct", func(t *testing.T) {
		t.Run("FailedToDecodeField", func(t *testing.T) {
			grp := envx.NewGroup("TEST")
			dec := envx.NewDecoder(grp)
			pos := envx.NewPathWalker()

			pos.Enter("MustFailed")
			grp.Add(envx.NewVar("MustFailed", "any"))

			err := dec.Decode(&struct {
				Time       time.Time
				MustFailed MustFailedArshaler
			}{
				MustFailed: MustFailedArshaler{},
			})
			Expect(t, err, IsCodeError(envx.CODE__DEC_FAILED_UNMARSHAL))
		})
	})
}
