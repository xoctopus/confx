package mq_test

import (
	"testing"

	"github.com/xoctopus/confx/pkg/types/mq"
)

func BenchmarkHashers(b *testing.B) {
	b.Run("fnv", func(b *testing.B) {
		h := mq.Fnv
		b.ResetTimer()
		for b.Loop() {
			_ = h("any")
			_ = h("")
			_ = h("little______________________________________long")
		}
	})

	b.Run("crc", func(b *testing.B) {
		h := mq.CRC
		b.ResetTimer()
		for b.Loop() {
			_ = h("any")
			_ = h("")
			_ = h("little______________________________________long")
		}
	})
}

func TestHashers(t *testing.T) {
	_ = mq.Fnv("any")
	_ = mq.Fnv("")
	_ = mq.Fnv("little______________________________________long")

	_ = mq.CRC("any")
	_ = mq.CRC("")
	_ = mq.CRC("little______________________________________long")
}
