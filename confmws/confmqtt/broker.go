package confmqtt

import (
	"log"
	"os"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/xoctopus/datatypex"
	"github.com/xoctopus/x/misc/retry"

	"github.com/xoctopus/confx/confmws/conftls"
)

func init() {
	mqtt.ERROR = log.New(os.Stdout, "[MQTT ERR]", log.LstdFlags)
	mqtt.CRITICAL = log.New(os.Stdout, "[MQTT CRI]", log.LstdFlags)
	mqtt.WARN = log.New(os.Stdout, "[MQTT WRN]", log.LstdFlags)
}

type Broker struct {
	Server        datatypex.Endpoint
	Retry         retry.Retry
	Timeout       time.Duration
	Keepalive     time.Duration
	RetainPublish bool
	QoS           QOS
	Cert          *conftls.X509KeyPair

	clients sync.Map
}

func (b *Broker) SetDefault() {
	b.Retry.SetDefault()
	if b.Keepalive == 0 {
		b.Keepalive = 3 * time.Second
	}
	if b.Timeout == 0 {
		b.Timeout = 10 * time.Second
	}
	if b.Server.IsZero() {
		b.Server.Host, b.Server.Port = "127.0.0.1", 1883
	}
	if b.Server.Scheme == "" {
		b.Server.Scheme = "tcp"
	}
	if b.QoS > QOS__ONLY_ONCE || b.QoS < 0 {
		b.QoS = QOS__ONCE
	}
}

func (b *Broker) Init() error {
	if b.Cert != nil {
		if err := b.Cert.Init(); err != nil {
			return err
		}
	}
	return b.Retry.Do(func() error {
		c, err := b.NewClient("", "ping")
		if err != nil {
			return err
		}
		b.Close(c)
		return nil
	})
}

func (b *Broker) options(cid string) *mqtt.ClientOptions {
	options := mqtt.NewClientOptions()
	if cid == "" {
		cid = uuid.NewString()
	}
	options.SetClientID(cid)

	if b.Server.Scheme == "ssl" {
		options.SetTLSConfig(b.Cert.Config())
	}

	options = options.AddBroker(b.Server.String())
	if b.Server.Username != "" {
		options.SetUsername(b.Server.Username)
		if b.Server.Password != "" {
			options.SetPassword(b.Server.Password.String())
		}
	}

	options.SetAutoReconnect(true)
	options.SetConnectRetryInterval(b.Timeout)

	// client resuming subscribing the topic before lost connection
	options.SetCleanSession(false)
	options.SetResumeSubs(true)

	// keepalive is the interval to probe if connection is alive
	options.SetKeepAlive(b.Keepalive)

	options.SetWriteTimeout(b.Timeout)
	options.SetConnectTimeout(b.Timeout)
	options.SetPingTimeout(b.Timeout)

	options.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		opt := c.OptionsReader()
		mqtt.WARN.Printf("connection lost: `%s` caused by %v", opt.ClientID(), err)
	})
	options.SetReconnectingHandler(func(_ mqtt.Client, opt *mqtt.ClientOptions) {
		mqtt.WARN.Printf("reconnecting: `%s`", opt.ClientID)
	})
	options.SetOnConnectHandler(func(c mqtt.Client) {
		opt := c.OptionsReader()
		mqtt.WARN.Printf("client connected: `%s`", opt.ClientID())
	})

	return options
}

func (b *Broker) Name() string {
	return "mqtt-broker"
}

func (b *Broker) LivenessCheck() map[string]string {
	m := map[string]string{}
	c, err := b.NewClient("", "liveness")
	if err != nil {
		m[b.Server.Hostname()] = err.Error()
		return m
	}
	b.Close(c)
	m[b.Server.Hostname()] = "ok"
	return m
}

func (b *Broker) NewClient(id, topic string) (*Client, error) {
	if topic == "" {
		return nil, ErrInvalidTopic
	}
	option := b.options(id)

	c := &Client{
		id:      id,
		qos:     b.QoS,
		retain:  b.RetainPublish,
		timeout: b.Timeout,
		topic:   topic,
		cli:     mqtt.NewClient(option),
	}
	if err := c.connect(); err != nil {
		return nil, err
	}
	b.clients.Store(id, c)
	return c, nil
}

func (b *Broker) Close(c *Client) {
	b.CloseByClientID(c.id)
}

func (b *Broker) CloseByClientID(id string) {
	if c, ok := b.clients.LoadAndDelete(id); ok && c != nil {
		cc := c.(*Client)
		cc.cli.Unsubscribe(cc.topic)
		cc.cli.Disconnect(100)
	}
}
