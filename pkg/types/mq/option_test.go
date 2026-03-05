package mq_test

import (
	"testing"

	"cgtech.gitlab.com/saitox/x/testx"

	"cgtech.gitlab.com/saitox/confx/pkg/types/mq"
)

func BenchmarkHasher(b *testing.B) {
	for name, f := range map[string]mq.Hasher{
		"fnv": mq.Fnv,
		"crc": mq.CRC,
	} {
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				_ = f("any")
				_ = f("")
				_ = f("too________________________________________long")
			}
		})
	}
}

func TestHasher(t *testing.T) {
	t.Log(mq.Fnv("any"))
	t.Log(mq.Fnv(""))
	t.Log(mq.Fnv("too________________________________________long"))

	t.Log(mq.CRC("any"))
	t.Log(mq.CRC(""))
	t.Log(mq.CRC("too________________________________________long"))
}

type MockOption struct {
	v any
}

func (*MockOption) OptionScheme() string {
	return "testing"
}

func WithV(o mq.Option) {
	if x, ok := o.(*MockOption); ok {
		x.v = 1
	}
}

func TestOptionApplyFunc_Apply(t *testing.T) {
	o := &MockOption{v: 0}
	mq.OptionApplyFunc(WithV).Apply(o)
	testx.Expect(t, o.v, testx.Equal[any](1))
}
