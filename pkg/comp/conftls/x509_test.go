package conftls_test

import (
	"os"
	"testing"

	"github.com/xoctopus/x/misc/must"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/comp/conftls"
)

var (
	key = string(must.NoErrorV(os.ReadFile("testdata/server.key")))
	crt = string(must.NoErrorV(os.ReadFile("testdata/server.crt")))
	ca  = string(must.NoErrorV(os.ReadFile("testdata/ca.crt")))
)

func TestX509KeyPair(t *testing.T) {
	t.Run("EmptyConfig", func(t *testing.T) {
		keypair := &conftls.X509KeyPair{}
		Expect(t, keypair.Init(), Succeed())
		Expect(t, keypair.IsZero(), BeTrue())
		Expect(t, keypair.Config(), Equal(conftls.DefaultTLSConfig))
	})

	t.Run("LoadFromPath", func(t *testing.T) {
		keypair := &conftls.X509KeyPair{
			Key: "testdata/server.key",
			Crt: "testdata/server.crt",
			CA:  "testdata/ca.crt",
		}
		Expect(t, keypair.Init(), Succeed())
		Expect(t, keypair.IsZero(), BeFalse())
		Expect(t, keypair.Config(), Equal(conftls.DefaultTLSConfig))
	})

	t.Run("FailedToLoadFromPath", func(t *testing.T) {
		t.Run("CA", func(t *testing.T) {
			keypair := &conftls.X509KeyPair{
				Key: "testdata/server.key",
				Crt: "testdata/server.crt",
				CA:  ca,
			}
			Expect(t, keypair.Init(), Failed())
			Expect(t, keypair.IsZero(), BeFalse())
			Expect(t, keypair.Config(), Equal(conftls.DefaultTLSConfig))
		})
		t.Run("Crt", func(t *testing.T) {
			keypair := &conftls.X509KeyPair{
				Key: "testdata/server.key",
				Crt: crt,
				CA:  ca,
			}
			Expect(t, keypair.Init(), Failed())
			Expect(t, keypair.IsZero(), BeFalse())
			Expect(t, keypair.Config(), Equal(conftls.DefaultTLSConfig))
		})
		t.Run("Key", func(t *testing.T) {
			keypair := &conftls.X509KeyPair{
				Key: key,
				Crt: crt,
				CA:  ca,
			}
			Expect(t, keypair.Init(), Failed())
			Expect(t, keypair.IsZero(), BeFalse())
			Expect(t, keypair.Config(), Equal(conftls.DefaultTLSConfig))
		})
	})

	t.Run("FailedToAppendCert", func(t *testing.T) {
		keypair := &conftls.X509KeyPair{
			Key: key,
			Crt: crt,
			CA:  "invalid ca",
		}
		Expect(t, keypair.Init(), Failed())
		Expect(t, keypair.IsZero(), BeFalse())
		Expect(t, keypair.Config(), Equal(conftls.DefaultTLSConfig))
	})
}
