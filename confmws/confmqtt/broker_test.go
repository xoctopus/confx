package confmqtt_test

import (
	"encoding/json"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	. "github.com/onsi/gomega"
	"github.com/xoctopus/x/misc/must"

	. "github.com/xoctopus/confx/confmws/confmqtt"
	"github.com/xoctopus/confx/confmws/conftls"
)

func NewMessage(msg string) *Message {
	return &Message{
		EventID:      uuid.New().String(),
		PubTimestamp: time.Now().UnixMilli(),
		Message:      msg,
	}
}

type Message struct {
	EventID      string
	PubTimestamp int64
	Message      string
}

var (
	topic  = "test_demo"
	broker = &Broker{QoS: QOS_UNKNOWN}
)

func TestBroker(t *testing.T) {
	NewWithT(t).Expect(broker.QoS).To(Equal(QOS_UNKNOWN))
	broker.SetDefault()
	NewWithT(t).Expect(broker.QoS).To(Equal(QOS__ONCE))

	err := broker.Server.UnmarshalText([]byte("ssl://name:passwd@broker.emqx.io:8883"))
	NewWithT(t).Expect(err).To(BeNil())

	broker.Cert = &conftls.X509KeyPair{}
	err = broker.Init()
	if err != nil {
		t.Skipf("failed to connect public mqtt broker: %s", broker.Server)
	}

	puber, err := broker.NewClient("pub", topic)
	NewWithT(t).Expect(err).To(BeNil())
	NewWithT(t).Expect(puber.ID()).To(Equal("pub"))

	suber, err := broker.NewClient("sub", topic)
	NewWithT(t).Expect(err).To(BeNil())
	NewWithT(t).Expect(suber.ID()).To(Equal("sub"))

	defer func() {
		err = puber.Unsubscribe()
		NewWithT(t).Expect(err).To(BeNil())
		err = suber.Unsubscribe()
		NewWithT(t).Expect(err).To(BeNil())
		broker.Close(puber)
		broker.Close(suber)
	}()

	err = suber.Subscribe(func(cli mqtt.Client, msg mqtt.Message) {
		pl := &Message{}
		ts := time.Now()
		NewWithT(t).Expect(json.Unmarshal(msg.Payload(), pl)).To(BeNil())
		t.Logf("topic: %s cost: %dms", msg.Topic(), ts.UnixMilli()-pl.PubTimestamp)
	})
	NewWithT(t).Expect(err).To(BeNil())

	num := 5
	for i := 0; i < num; i++ {
		err = puber.Publish(must.NoErrorV(json.Marshal(NewMessage("payload"))))
		if err != nil {
			t.Skip("failed to publish message")
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Log(broker.LivenessCheck())
}

func TestBrokerExt(t *testing.T) {
	t.Run("InvalidCert", func(t *testing.T) {
		b := &Broker{
			Cert: &conftls.X509KeyPair{
				Key: "invalid",
				Crt: "invalid",
				CA:  "invalid",
			},
		}
		NewWithT(t).Expect(b.Init()).NotTo(BeNil())
	})
	t.Run("InvalidEndpoint", func(t *testing.T) {
		b := &Broker{}
		b.Retry.Repeats = -1
		b.SetDefault()
		NewWithT(t).Expect(b.Init()).NotTo(BeNil())
		liveness := b.LivenessCheck()[b.Server.Hostname()]
		NewWithT(t).Expect(liveness).NotTo(Equal("ok"))
		NewWithT(t).Expect(b.Name()).To(Equal("mqtt-broker"))
	})
	t.Run("InvalidTopic", func(t *testing.T) {
		c, err := (&Broker{}).NewClient("", "")
		NewWithT(t).Expect(err).To(Equal(ErrInvalidTopic))
		NewWithT(t).Expect(c).To(BeNil())
	})
}
