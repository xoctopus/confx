package conftls_test

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/xoctopus/x/misc/must"

	"github.com/xoctopus/confx/confmws/conftls"
)

var (
	key  = string(must.NoErrorV(os.ReadFile("testdata/client/key")))
	cert = string(must.NoErrorV(os.ReadFile("testdata/client/cert")))
	ca   = string(must.NoErrorV(os.ReadFile("testdata/ca.cert")))
)

func TestX509KeyPair(t *testing.T) {
	t.Run("EmptyConfig", func(t *testing.T) {
		keypair := &conftls.X509KeyPair{}
		NewWithT(t).Expect(keypair.Init()).To(BeNil())
		NewWithT(t).Expect(keypair.IsZero()).To(BeTrue())
		NewWithT(t).Expect(keypair.Config()).To(Equal(conftls.DefaultTLSConfig))
	})

	t.Run("LoadFromPath", func(t *testing.T) {
		keypair := &conftls.X509KeyPair{
			Key: "testdata/client/key",
			Crt: "testdata/client/cert",
			CA:  "testdata/ca.cert",
		}
		NewWithT(t).Expect(keypair.Init()).To(BeNil())
		NewWithT(t).Expect(keypair.IsZero()).To(BeFalse())
		NewWithT(t).Expect(keypair.Config()).NotTo(Equal(conftls.DefaultTLSConfig))
	})

	t.Run("FailedToLoadFromPath", func(t *testing.T) {
		t.Run("CA", func(t *testing.T) {
			keypair := &conftls.X509KeyPair{
				Key: "testdata/client/key",
				Crt: "testdata/client/cert",
				CA:  ca,
			}
			NewWithT(t).Expect(keypair.Init()).NotTo(BeNil())
			NewWithT(t).Expect(keypair.IsZero()).To(BeFalse())
			NewWithT(t).Expect(keypair.Config()).To(Equal(conftls.DefaultTLSConfig))
		})
		t.Run("Crt", func(t *testing.T) {
			keypair := &conftls.X509KeyPair{
				Key: "testdata/client/key",
				Crt: cert,
				CA:  ca,
			}
			NewWithT(t).Expect(keypair.Init()).NotTo(BeNil())
			NewWithT(t).Expect(keypair.IsZero()).To(BeFalse())
			NewWithT(t).Expect(keypair.Config()).To(Equal(conftls.DefaultTLSConfig))
		})
		t.Run("Key", func(t *testing.T) {
			keypair := &conftls.X509KeyPair{
				Key: key,
				Crt: cert,
				CA:  ca,
			}
			NewWithT(t).Expect(keypair.Init()).To(BeNil())
			NewWithT(t).Expect(keypair.IsZero()).To(BeFalse())
			NewWithT(t).Expect(keypair.Config()).NotTo(Equal(conftls.DefaultTLSConfig))
		})
	})

	t.Run("FailedToAppendCert", func(t *testing.T) {
		keypair := &conftls.X509KeyPair{
			Key: key,
			Crt: cert,
			CA:  "invalid ca",
		}
		NewWithT(t).Expect(keypair.Init()).NotTo(BeNil())
		NewWithT(t).Expect(keypair.IsZero()).To(BeFalse())
		NewWithT(t).Expect(keypair.Config()).To(Equal(conftls.DefaultTLSConfig))
	})
}
