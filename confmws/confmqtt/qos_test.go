package confmqtt_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/xoctopus/confx/confmws/confmqtt"
)

func TestQOS(t *testing.T) {
	for _, s := range []string{"ONCE", "AT_LEAST_ONCE", "ONLY_ONCE", "Invalid", ""} {
		qos := confmqtt.QOS_UNKNOWN
		if err := qos.UnmarshalText([]byte(s)); err != nil {
			NewWithT(t).Expect(err).To(Equal(confmqtt.InvalidQOS))
		} else {
			NewWithT(t).Expect(qos.String()).To(Equal(s))
		}
	}
	for _, l := range []string{"0", "1", "2", "Invalid", ""} {
		if qos, err := confmqtt.ParseQOSFromLabel(l); err != nil {
			NewWithT(t).Expect(err).To(Equal(confmqtt.InvalidQOS))
		} else {
			NewWithT(t).Expect(qos.Label()).To(Equal(l))
		}
	}
	for _, qos := range []confmqtt.QOS{-1, 0, 1, 2, 10} {
		data, err := qos.MarshalText()
		if err != nil {
			NewWithT(t).Expect(err).To(Equal(confmqtt.InvalidQOS))
			NewWithT(t).Expect(qos.Label()).To(Equal("UNKNOWN"))
		} else {
			NewWithT(t).Expect(qos.String()).To(Equal(string(data)))
			NewWithT(t).Expect(qos.Int()).To(Equal(int(qos)))
		}
	}
}
