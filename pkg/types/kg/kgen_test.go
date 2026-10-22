package kg_test

import (
	"strings"
	"testing"
	_ "unsafe"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/types/kg"
)

//go:linkname reset github.com/xoctopus/confx/pkg/types/kg.reset
func reset(_ *kg.KeyGen)

const sep = kg.SegmentSeparator

func key(parts ...string) string {
	return strings.Join(parts, sep)
}

func TestSanitizeSegment(t *testing.T) {
	Expect(t, kg.SanitizeSegment("  a b  "), Equal("a_b"))
	Expect(t, kg.SanitizeSegment("10:50"), Equal("10_50"))
	Expect(t, kg.SanitizeSegment("pod-1"), Equal("pod_1"))
	Expect(t, kg.SanitizeSegment("   "), Equal(""))
}

func TestNormalizeCreator(t *testing.T) {
	t.Run("sanitize and upper", func(t *testing.T) {
		Expect(t, kg.NormalizeCreator("svc-alpha"), Equal("SVC_ALPHA"))
		Expect(t, kg.NormalizeCreator("svc alpha"), Equal("SVC_ALPHA"))
		Expect(t, kg.NormalizeCreator("svc:alpha"), Equal("SVC_ALPHA"))
	})

	t.Run("empty after normalize panics", func(t *testing.T) {
		ExpectPanic[error](t, func() { _ = kg.NormalizeCreator("") })
		ExpectPanic[error](t, func() { _ = kg.NormalizeCreator("   ") })
	})

	t.Run("reserved panics", func(t *testing.T) {
		ExpectPanic[error](t, func() { _ = kg.NormalizeCreator("PEER") })
		ExpectPanic[error](t, func() { _ = kg.NormalizeCreator("ANY") })
	})
}

func TestNormalizeAudience(t *testing.T) {
	Expect(t, kg.NormalizeAudience("svc-gamma"), Equal("SVC_GAMMA"))
	ExpectPanic[error](t, func() { _ = kg.NormalizeAudience("") })
	ExpectPanic[error](t, func() { _ = kg.NormalizeAudience("PEER") })
}

func TestKeyGen(t *testing.T) {
	var g kg.KeyGen
	t.Cleanup(func() { reset(&g) })
	g.Init("svc-alpha")

	Expect(t, g.Creator(), Equal("SVC_ALPHA"))
	id := g.InstanceID()
	Expect(t, len(id), Equal(26))

	Expect(t, g.Key("SESSION", uint64(42)), Equal(key("SVC_ALPHA", id, "SESSION", "42")))
	Expect(t, g.SharedKey("SESSION", uint64(42)), Equal(key("SVC_ALPHA", "PEER", "SESSION", "42")))
	Expect(t, g.GlobalKey("REBOOTING", "93_1"), Equal(key("SVC_ALPHA", "ANY", "REBOOTING", "93_1")))
	Expect(t, g.ShareTo("svc-gamma", "REBOOTING", "93_1"),
		Equal(key("SVC_ALPHA", "SVC_GAMMA", "REBOOTING", "93_1")))

	Expect(t, g.SharedKeyFrom("svc-beta", "SESSION", uint64(1)),
		Equal(key("SVC_BETA", "PEER", "SESSION", "1")))

	g.Init("other")
	Expect(t, g.Creator(), Equal("SVC_ALPHA"))
	Expect(t, g.InstanceID(), Equal(id))
}

func TestKeyGenWithPrefix(t *testing.T) {
	var g kg.KeyGen
	t.Cleanup(func() { reset(&g) })
	g.Init("svc-alpha", kg.WithPrefix("1050"))

	Expect(t, g.Prefix(), Equal("1050"))
	Expect(t, g.SharedKey("SESSION", uint64(1)), Equal(key("1050", "SVC_ALPHA", "PEER", "SESSION", "1")))

	var g2 kg.KeyGen
	t.Cleanup(func() { reset(&g2) })
	g2.Init("svc-alpha", kg.WithPrefix("  "))
	Expect(t, g2.Prefix(), Equal(""))
	Expect(t, g2.SharedKey("SESSION", uint64(1)), Equal(key("SVC_ALPHA", "PEER", "SESSION", "1")))

	var g3 kg.KeyGen
	t.Cleanup(func() { reset(&g3) })
	g3.Init("svc-alpha", kg.WithPrefix("10 50"))
	Expect(t, g3.Prefix(), Equal("10_50"))
}

func TestKeyGenWithInstanceID(t *testing.T) {
	const custom = "01ARZ3NDEKTSV4RRFFQ69G5FAV"

	var g kg.KeyGen
	t.Cleanup(func() { reset(&g) })
	g.Init("svc-alpha", kg.WithInstanceID(custom))
	Expect(t, g.Key("SESSION", uint64(42)), Equal(key("SVC_ALPHA", custom, "SESSION", "42")))
	Expect(t, g.SharedKey("SESSION", uint64(42)), Equal(key("SVC_ALPHA", "PEER", "SESSION", "42")))

	var g2 kg.KeyGen
	t.Cleanup(func() { reset(&g2) })
	g2.Init("svc-alpha", kg.WithPrefix("1050"), kg.WithInstanceID("pod-1"))
	Expect(t, g2.Key("CACHE", "x"), Equal(key("1050", "SVC_ALPHA", "pod_1", "CACHE", "x")))

	var g3 kg.KeyGen
	t.Cleanup(func() { reset(&g3) })
	g3.Init("svc-alpha", kg.WithInstanceID("   "))
	Expect(t, len(g3.InstanceID()), Equal(26))

	var g4 kg.KeyGen
	t.Cleanup(func() { reset(&g4) })
	g4.Init("svc-alpha", kg.WithInstanceID("pod 1"))
	Expect(t, g4.InstanceID(), Equal("pod_1"))
}

func TestKeyGenSanitizeDomainAndBiz(t *testing.T) {
	var g kg.KeyGen
	t.Cleanup(func() { reset(&g) })
	g.Init("svc-alpha")

	Expect(t, g.SharedKey("SES SION", uint64(1)), Equal(key("SVC_ALPHA", "PEER", "SES_SION", "1")))
	Expect(t, g.SharedKey("SES:SION", "93 1"), Equal(key("SVC_ALPHA", "PEER", "SES_SION", "93_1")))
}

func TestKeyGenEmptySegmentPanics(t *testing.T) {
	var g kg.KeyGen
	t.Cleanup(func() { reset(&g) })
	g.Init("svc-alpha")

	ExpectPanic[error](t, func() { _ = g.SharedKey("   ", uint64(1)) })
	ExpectPanic[error](t, func() { _ = g.SharedKey("SESSION", "  ") })

	var g2 kg.KeyGen
	t.Cleanup(func() { reset(&g2) })
	ExpectPanic[error](t, func() { g2.Init("PEER") })
}
