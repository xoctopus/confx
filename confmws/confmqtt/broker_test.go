package confmqtt_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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
	NewWithT(t).Expect(puber.Topic()).To(Equal(topic))

	suber, err := broker.NewClient("sub", topic)
	NewWithT(t).Expect(err).To(BeNil())
	NewWithT(t).Expect(suber.ID()).To(Equal("sub"))
	NewWithT(t).Expect(suber.Topic()).To(Equal(topic))

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

func TestClientReconnection(t *testing.T) {
	b := &Broker{}
	err := b.Server.UnmarshalText([]byte("tcp://broker.emqx.io:1883"))
	NewWithT(t).Expect(err).To(BeNil())

	b.SetDefault()
	b.QoS = QOS__ONLY_ONCE // make clients will not lose message

	b.Cert = &conftls.X509KeyPair{}
	err = b.Init()
	if err != nil {
		t.Skipf("failed to connect public mqtt broker: %s", b.Server)
	}

	cpub, err := b.NewClient("pub__reconnection", "test_reconnection")
	if err != nil {
		t.Skipf("failed to new pub client: %s", b.Server)
	}
	csub, err := b.NewClient("sub__reconnection", "test_reconnection")
	if err != nil {

	}

	sig := make(chan struct{})
	lmt := 10

	// start subscribing until received limit or sequence great equal than limit.
	subed := make([]int, 0)
	err = csub.Subscribe(func(_ mqtt.Client, m mqtt.Message) {
		msg := string(m.Payload())
		seq, err := strconv.Atoi(strings.Split(msg, "-")[0])
		if err != nil {
			t.Error("subscribed unexpected message")
			sig <- struct{}{}
			return
		}
		subed = append(subed, seq)
		t.Logf("sub %s", msg)
		if len(subed) >= lmt && seq >= lmt {
			time.Sleep(time.Second)
			b.Close(csub)
			sig <- struct{}{}
		}
	})

	if err != nil {
		t.Error("failed to subscribe")
		sig <- struct{}{}
	}

	// start publishing until sequence great equal than limit.
	pubed := make([]int, 0)
	go func() {
		seq := 0
		for {
			if seq > lmt {
				b.Close(cpub)
				return
			}
			msg := fmt.Sprintf("%d-%d", seq, time.Now().UnixNano())
			err := cpub.Publish(msg)
			if err != nil {
				t.Logf("failed to publish seq: %d %v", seq, err)
				goto TryLater
			}
			pubed = append(pubed, seq)
			t.Logf("pub %s", msg)
			seq++
		TryLater:
			time.Sleep(time.Second)
		}
	}()

	<-sig
	t.Logf("test finished ")
	t.Logf("published:  %v", pubed)
	t.Logf("subscribed: %v", subed)
}
