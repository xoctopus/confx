package kg_test

import (
	"fmt"
	"testing"
	_ "unsafe"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/types/kg"
)

//go:linkname reset github.com/xoctopus/confx/pkg/types/kg.reset
func reset(g *kg.KeyGen)

func TestNormalizeCreator(t *testing.T) {
	t.Run("to upper and dash to underscore", func(t *testing.T) {
		Expect(t, kg.NormalizeCreator("svc-alpha"), Equal("SVC_ALPHA"))
		Expect(t, kg.NormalizeCreator("  svc-beta  "), Equal("SVC_BETA"))
	})

	t.Run("empty panics", func(t *testing.T) {
		ExpectPanic[error](t, func() { _ = kg.NormalizeCreator("") })
		ExpectPanic[error](t, func() { _ = kg.NormalizeCreator("   ") })
	})

	t.Run("reserved panics", func(t *testing.T) {
		ExpectPanic[error](t, func() { _ = kg.NormalizeCreator("PEER") })
		ExpectPanic[error](t, func() { _ = kg.NormalizeCreator("peer") })
		ExpectPanic[error](t, func() { _ = kg.NormalizeCreator("ANY") })
		ExpectPanic[error](t, func() { _ = kg.NormalizeCreator("any") })
	})
}

func TestNormalizeAudience(t *testing.T) {
	t.Run("to upper and dash to underscore", func(t *testing.T) {
		Expect(t, kg.NormalizeAudience("svc-gamma"), Equal("SVC_GAMMA"))
	})

	t.Run("empty panics", func(t *testing.T) {
		ExpectPanic[error](t, func() { _ = kg.NormalizeAudience("") })
	})

	t.Run("reserved panics", func(t *testing.T) {
		ExpectPanic[error](t, func() { _ = kg.NormalizeAudience("PEER") })
		ExpectPanic[error](t, func() { _ = kg.NormalizeAudience("ANY") })
	})
}

func TestKeyGen(t *testing.T) {
	var g kg.KeyGen
	t.Cleanup(func() { reset(&g) })

	g.Init("svc-alpha")
	Expect(t, g.Creator(), Equal("SVC_ALPHA"))
	Expect(t, g.InstanceID(), NotEqual(""))
	Expect(t, len(g.InstanceID()), Equal(26))

	id := g.InstanceID()

	t.Run("Key uses instance ulid", func(t *testing.T) {
		Expect(t, g.Key("SESSION", int64(42)),
			Equal(fmt.Sprintf("SVC_ALPHA:%s:SESSION:42", id)))
	})

	t.Run("SharedKey uses PEER", func(t *testing.T) {
		Expect(t, g.SharedKey("SESSION", int64(42)),
			Equal("SVC_ALPHA:PEER:SESSION:42"))
	})

	t.Run("GlobalKey uses ANY", func(t *testing.T) {
		Expect(t, g.GlobalKey("REBOOT_BET", "93_1"),
			Equal("SVC_ALPHA:ANY:REBOOT_BET:93_1"))
	})

	t.Run("ShareTo normalizes audience", func(t *testing.T) {
		Expect(t, g.ShareTo("svc-gamma", "REBOOT_BET", "93_1"),
			Equal("SVC_ALPHA:SVC_GAMMA:REBOOT_BET:93_1"))
	})

	t.Run("From variants", func(t *testing.T) {
		Expect(t, g.SharedKeyFrom("svc-beta", "SESSION", int64(1)),
			Equal("SVC_BETA:PEER:SESSION:1"))
		Expect(t, g.GlobalKeyFrom("svc-beta", "REBOOT_BET", "93_1"),
			Equal("SVC_BETA:ANY:REBOOT_BET:93_1"))
		Expect(t, g.ShareFrom("svc-beta", "svc-gamma", "REBOOT_BET", "93_1"),
			Equal("SVC_BETA:SVC_GAMMA:REBOOT_BET:93_1"))
	})

	t.Run("Init again is no-op", func(t *testing.T) {
		g.Init("other-service")
		Expect(t, g.Creator(), Equal("SVC_ALPHA"))
		Expect(t, g.InstanceID(), Equal(id))
	})
}

func TestKeyGenWithPrefix(t *testing.T) {
	var g kg.KeyGen
	t.Cleanup(func() { reset(&g) })

	g.Init("svc-alpha", kg.WithPrefix("1050"))
	Expect(t, g.Creator(), Equal("SVC_ALPHA"))
	Expect(t, g.Prefix(), Equal("1050"))

	t.Run("SharedKey with prefix", func(t *testing.T) {
		Expect(t, g.SharedKey("SESSION", int64(42)),
			Equal("1050:SVC_ALPHA:PEER:SESSION:42"))
	})

	t.Run("GlobalKey with prefix", func(t *testing.T) {
		Expect(t, g.GlobalKey("REBOOT_BET", "93_1"),
			Equal("1050:SVC_ALPHA:ANY:REBOOT_BET:93_1"))
	})
}

func TestKeyGenReservedCreator(t *testing.T) {
	var g kg.KeyGen
	t.Cleanup(func() { reset(&g) })

	ExpectPanic[error](t, func() { g.Init("PEER") })
}
