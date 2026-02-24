package mq_test

import (
	"testing"

	"github.com/xoctopus/confx/pkg/types/mq"
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
