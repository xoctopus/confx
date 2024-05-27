package envconf_test

import (
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/xoctopus/datatypex"

	"github.com/xoctopus/confx/envconf"
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

			grp.Reset()
			err = enc.Encode(map[string]string(nil))
			NewWithT(t).Expect(err).To(BeNil())
		})
	})

	t.Run("Slice", func(t *testing.T) {
		grp.Reset()

		err := enc.Encode([]int{1, 2, 3})
		NewWithT(t).Expect(err).To(BeNil())

		err = enc.Encode([]MustFailed{{}})
		NewWithT(t).Expect(err).NotTo(BeNil())
		NewWithT(t).Expect(errors.Is(err, ErrMustMarshalFailed)).To(BeTrue())
	})

	t.Run("Struct", func(t *testing.T) {
		grp.Reset()

		val := &struct {
			unexported any
			PtrIsNil   *int
			HasTag     *datatypex.Address `env:"address,optional,upstream,copy,expose"`
			SkipTag    datatypex.Endpoint `env:"-"`
			Inline
			datatypex.Password
			// todo if anonymous field implements TextMarshaller will overwrite?
			MustFailed MustFailed
		}{
			HasTag: datatypex.NewAddress("group", "filename.png"),
			Inline: Inline{
				String: "inline string",
				Int:    100,
			},
			Password: datatypex.Password("password"),
		}
		err := enc.Encode(val)
		NewWithT(t).Expect(err).NotTo(BeNil())
		NewWithT(t).Expect(errors.Is(err, ErrMustMarshalFailed)).To(BeTrue())

		v := grp.Get("address")
		NewWithT(t).Expect(v).NotTo(BeNil())
		NewWithT(t).Expect(v.Optional).To(BeTrue())
	})
}
