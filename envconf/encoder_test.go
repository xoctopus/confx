package envconf_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/xoctopus/confx/envconf"
)

func TestEncoder_Encode(t *testing.T) {
	t.Run("Invalid", func(t *testing.T) {
		t.Run("InvalidGroupName", func(t *testing.T) {
			defer func() {
				err := recover().(string)
				NewWithT(t).Expect(err).To(HavePrefix("invalid group name"))
			}()
			envconf.NewGroup("x.x")
		})
		grp := envconf.NewGroup("TEST")
		enc := envconf.NewEncoder(grp)

		err := enc.Encode(nil)
		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(grp.Len()).To(Equal(0))
	})

	t.Run("IgnoredKinds", func(t *testing.T) {
		grp := envconf.NewGroup("TEST")
		enc := envconf.NewEncoder(grp)

		err := enc.Encode(&struct {
			Func func()
			Chan chan struct{}
			Any  any
		}{})
		NewWithT(t).Expect(err).To(BeNil())
	})

	t.Run("Map", func(t *testing.T) {
		t.Run("UnexpectMapKeyType", func(t *testing.T) {
			grp := envconf.NewGroup("TEST")
			enc := envconf.NewEncoder(grp)
			pos := envconf.NewPathWalker()

			type Key struct{}
			target := enc.Encode(map[Key]string{})
			expect := envconf.NewError(pos, envconf.E_ENC__INVALID_MAP_KEY_TYPE)
			NewWithT(t).Expect(errors.Is(expect, target)).To(BeTrue())
		})
		t.Run("UnexpectMapKeyValue", func(t *testing.T) {
			grp := envconf.NewGroup("TEST")
			enc := envconf.NewEncoder(grp)
			pos := envconf.NewPathWalker()
			expect := envconf.NewError(pos, envconf.E_ENC__INVALID_MAP_KEY_VALUE)
			t.Run("InvalidInteger", func(t *testing.T) {
				target := enc.Encode(map[int]string{-1: "any"})
				NewWithT(t).Expect(errors.Is(expect, target)).To(BeTrue())
			})
			t.Run("InvalidString", func(t *testing.T) {
				target := enc.Encode(map[string]string{"": "any"})
				NewWithT(t).Expect(errors.Is(expect, target)).To(BeTrue())

				target = enc.Encode(map[string]string{"a b c": "any"})
				NewWithT(t).Expect(errors.Is(expect, target)).To(BeTrue())
			})
		})
		t.Run("FailedToMarshal", func(t *testing.T) {
			grp := envconf.NewGroup("TEST")
			enc := envconf.NewEncoder(grp)
			pos := envconf.NewPathWalker()

			pos.Enter(0)
			expect := envconf.NewError(pos, envconf.E_ENC__FAILED_MARSHAL)

			target := enc.Encode(map[int]MustFailedArshaler{0: {}})
			NewWithT(t).Expect(errors.Is(expect, target)).To(BeTrue())
		})

		t.Run("Success", func(t *testing.T) {
			t.Run("StringKey", func(t *testing.T) {
				grp := envconf.NewGroup("TEST")
				enc := envconf.NewEncoder(grp)
				val := map[string]int{"1": 1, "2": 2}

				err := enc.Encode(val)
				NewWithT(t).Expect(err).To(BeNil())
				NewWithT(t).Expect(grp.Len()).To(Equal(2))
				NewWithT(t).Expect(grp.Len()).To(Equal(len(val)))
				for k := range val {
					v := grp.Get(k)
					NewWithT(t).Expect(fmt.Sprint(val[k])).To(Equal(v.Value()))
					NewWithT(t).Expect(fmt.Sprint(k)).To(Equal(v.Key()))
				}
			})
			t.Run("IntegerKey", func(t *testing.T) {
				grp := envconf.NewGroup("TEST")
				enc := envconf.NewEncoder(grp)
				val := map[int]string{1: "1", 2: "2"}

				err := enc.Encode(val)
				NewWithT(t).Expect(err).To(BeNil())
				NewWithT(t).Expect(grp.Len()).To(Equal(2))
				NewWithT(t).Expect(grp.Len()).To(Equal(len(val)))
				for k := range val {
					v := grp.Get(fmt.Sprint(k))
					NewWithT(t).Expect(fmt.Sprint(val[k])).To(Equal(v.Value()))
					NewWithT(t).Expect(fmt.Sprint(k)).To(Equal(v.Key()))
				}
			})
			t.Run("MapMap", func(t *testing.T) {
				grp := envconf.NewGroup("TEST")
				enc := envconf.NewEncoder(grp)
				val := map[int]map[string]int{
					1: {"1": 1},
					2: {"2": 2},
				}

				err := enc.Encode(val)
				NewWithT(t).Expect(err).To(BeNil())
				NewWithT(t).Expect(grp.Len()).To(Equal(2))
				NewWithT(t).Expect(grp.Len()).To(Equal(len(val)))
				NewWithT(t).Expect(grp.Get("1_1").Value()).To(Equal("1"))
				NewWithT(t).Expect(grp.Get("2_2").Value()).To(Equal("2"))
			})
			t.Run("NilMap", func(t *testing.T) {
				grp := envconf.NewGroup("TEST")
				enc := envconf.NewEncoder(grp)
				err := enc.Encode(map[int]int(nil))
				NewWithT(t).Expect(err).To(BeNil())
				NewWithT(t).Expect(grp.Len()).To(Equal(0))
			})
		})
	})

	t.Run("Slice", func(t *testing.T) {
		t.Run("Failed", func(t *testing.T) {
			grp := envconf.NewGroup("TEST")
			enc := envconf.NewEncoder(grp)

			pos := envconf.NewPathWalker()
			pos.Enter(0)

			expect := envconf.NewError(pos, envconf.E_ENC__FAILED_MARSHAL)
			target := enc.Encode([]MustFailedArshaler{{}})
			NewWithT(t).Expect(errors.Is(expect, target)).To(BeTrue())
		})
		grp := envconf.NewGroup("TEST")
		enc := envconf.NewEncoder(grp)

		err := enc.Encode([]int{1, 2, 3})
		NewWithT(t).Expect(err).To(BeNil())
	})

	t.Run("Struct", func(t *testing.T) {
		t.Run("FailedEncodeField", func(t *testing.T) {
			grp := envconf.NewGroup("TEST")
			enc := envconf.NewEncoder(grp)
			pos := envconf.NewPathWalker()
			pos.Enter("X")

			expect := envconf.NewError(pos, envconf.E_ENC__FAILED_MARSHAL)
			target := enc.Encode(struct {
				X MustFailedArshaler
			}{
				X: MustFailedArshaler{},
			})
			NewWithT(t).Expect(errors.Is(expect, target)).To(BeTrue())
		})
		t.Run("DuplicatedGroupKey", func(t *testing.T) {
			grp := envconf.NewGroup("TEST")
			enc := envconf.NewEncoder(grp)
			pos := envconf.NewPathWalker()

			pos.Enter("X")

			expect := envconf.NewError(pos, envconf.E_ENC__DUPLICATE_GROUP_KEY)
			target := enc.Encode(struct {
				NIL *int
				X   int `env:"x"`
				Y   int `env:"x"`
			}{})
			NewWithT(t).Expect(errors.Is(expect, target)).To(BeTrue())
		})
	})
}
