package envconf_test

import (
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/sincospro/datatypes"

	"github.com/sincospro/conf/envconf"
)

func TestEncoder_Encode(t *testing.T) {
	grp := envconf.NewGroup("TEST")
	enc := envconf.NewEncoder(grp)

	t.Run("Invalid", func(t *testing.T) {
		err := enc.Encode(nil)
		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(grp.Len()).To(Equal(0))
	})

	t.Run("SkippedKinds", func(t *testing.T) {
		err := enc.Encode(&struct {
			Func func()
			Chan chan struct{}
		}{})
		NewWithT(t).Expect(err).To(BeNil())
	})

	t.Run("Map", func(t *testing.T) {
		t.Run("UnexpectMapKeyType", func(t *testing.T) {
			type Key struct{}
			err := enc.Encode(map[Key]string{})
			NewWithT(t).Expect(err).NotTo(BeNil())
			NewWithT(t).Expect(envconf.NewErrMarshalUnexpectMapKeyType(reflect.TypeOf(Key{})).Is(err)).To(BeTrue())
			NewWithT(t).Expect(err.Error()).To(ContainSubstring("Key"))
		})
		t.Run("UnexpectMapKeyValue", func(t *testing.T) {
			err := enc.Encode(map[string]string{"": "any"})
			NewWithT(t).Expect(err).NotTo(BeNil())
			NewWithT(t).Expect(envconf.ErrUnexpectMapKeyValue("")).To(Equal(err))

			err = enc.Encode(map[string]string{"a b c": "any"})
			NewWithT(t).Expect(err).NotTo(BeNil())
			NewWithT(t).Expect(envconf.ErrUnexpectMapKeyValue("a b c")).To(Equal(err))
		})
		t.Run("FailedToMarshal", func(t *testing.T) {
			err := enc.Encode(map[int]MustFailed{0: {}})
			NewWithT(t).Expect(err).NotTo(BeNil())
			NewWithT(t).Expect(errors.Is(err, ErrMustMarshalFailed)).To(BeTrue())
		})
		t.Run("Success", func(t *testing.T) {
			grp.Reset()

			err := enc.Encode(map[string]string{
				"Key1": "value1",
				"Key2": "value2",
				"Key3": "value3",
			})
			NewWithT(t).Expect(err).To(BeNil())
			grp.Print()

			grp.Reset()
			err = enc.Encode(map[string]string(nil))
			NewWithT(t).Expect(err).To(BeNil())
		})
	})

	t.Run("Slice", func(t *testing.T) {
		grp.Reset()

		err := enc.Encode([]int{1, 2, 3})
		NewWithT(t).Expect(err).To(BeNil())
		grp.Print()

		err = enc.Encode([]MustFailed{{}})
		NewWithT(t).Expect(err).NotTo(BeNil())
		NewWithT(t).Expect(errors.Is(err, ErrMustMarshalFailed)).To(BeTrue())
	})

	t.Run("Struct", func(t *testing.T) {
		val := &struct {
			unexported any
			PtrIsNil   *int
			HasTag     *datatypes.Address `env:"address,opt,upstream,copy,expose"`
			SkipTag    datatypes.Endpoint `env:"-"`
			Inline
			datatypes.Password
			// MustFailed todo if inline implements TextMarshaller will overwrite?
			MustFailed MustFailed
		}{
			HasTag: &datatypes.Address{
				Group: "group",
				Key:   "key",
				Ext:   "ext",
			},
			Inline: Inline{
				String: "inline string",
				Int:    100,
			},
			Password: datatypes.Password("password"),
		}
		err := enc.Encode(val)
		NewWithT(t).Expect(err).NotTo(BeNil())
		NewWithT(t).Expect(errors.Is(err, ErrMustMarshalFailed)).To(BeTrue())

		grp.Print()
	})
}