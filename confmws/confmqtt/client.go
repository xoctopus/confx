package confmqtt

import (
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
)

type Client struct {
	id      string // id client id
	topic   string // topic registered topic
	qos     QOS    // qos should be 0, 1 or 2
	retain  bool
	timeout time.Duration
	cli     mqtt.Client
}

func (c *Client) ID() string {
	return c.id
}

func (c *Client) Topic() string {
	return c.topic
}

func (c *Client) WithTopic(topic string) *Client {
	cc := *c
	cc.topic = topic
	return &cc
}

func (c *Client) wait(tok mqtt.Token) error {
	tok.WaitTimeout(c.timeout)
	if err := tok.Error(); err != nil {
		return errors.Wrap(err, "failed to wait")
	}
	return nil
}

func (c *Client) connect() error {
	return c.wait(c.cli.Connect())
}

func (c *Client) Publish(payload interface{}) error {
	return c.wait(c.cli.Publish(c.topic, byte(c.qos), c.retain, payload))
}

func (c *Client) Subscribe(handler mqtt.MessageHandler) error {
	return c.wait(c.cli.Subscribe(c.topic, byte(c.qos), handler))
}

func (c *Client) Unsubscribe() error {
	return c.wait(c.cli.Unsubscribe(c.topic))
}
