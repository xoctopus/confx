package confmq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/xoctopus/sfid/pkg/sfid"
	"github.com/xoctopus/x/misc/must"
)

type Message interface {
	Topic() string
	ID() int64
	Timestamp() time.Time

	Data() []byte
	Extra() map[string][]string
}

type MessageArshaler interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

type MutMessage interface {
	Message

	AddExtra(string, string)
}

// ParseMessage data from message queue consumer
func ParseMessage(data []byte) (Message, error) {
	m := &message{}
	if err := m.Unmarshal(data); err != nil {
		return nil, err
	}
	return m, nil
}

func NewMessage(ctx context.Context, topic string, v any) Message {
	raw := must.NoErrorV(MarshalV(v))
	return NewMessageFromRaw(ctx, topic, raw)
}

func NewMessageFromRaw(ctx context.Context, topic string, raw []byte) Message {
	idg := sfid.Must(ctx)

	return &message{
		topic:     topic,
		id:        idg.ID(),
		data:      raw,
		timestamp: time.Now(),
		extra:     make(map[string][]string),
	}
}

type message struct {
	// meta
	topic     string
	id        int64
	timestamp time.Time

	// payload
	data []byte

	// extra
	extra map[string][]string
}

func (m *message) Topic() string {
	return m.topic
}

func (m *message) ID() int64 {
	return m.id
}

func (m *message) Timestamp() time.Time {
	return m.timestamp
}

func (m *message) Data() []byte {
	return m.data
}

func (m *message) Extra() map[string][]string {
	return m.extra
}

func (m *message) AddExtra(key, val string) {
	m.extra[key] = append(m.extra[key], val)
}

func (m message) Marshal() ([]byte, error) {
	return json.Marshal(newMeta(&m))
}

func (m *message) Unmarshal(data []byte) error {
	meta := &messageMeta{}
	err := json.Unmarshal(data, meta)
	if err != nil {
		return err
	}

	m.topic = meta.Topic
	m.id = meta.ID
	m.extra = meta.Extra
	m.data = meta.Data
	m.timestamp = meta.Timestamp
	m.extra = meta.Extra
	return nil
}

func newMeta(m Message) *messageMeta {
	return &messageMeta{
		Topic:     m.Topic(),
		ID:        m.ID(),
		Timestamp: m.Timestamp(),
		Data:      m.Data(),
		Extra:     m.Extra(),
	}
}

type messageMeta struct {
	Topic     string              `json:"topic"`
	ID        int64               `json:"id"`
	Timestamp time.Time           `json:"timestamp"`
	Data      []byte              `json:"data,omitempty"`
	Extra     map[string][]string `json:"extra,omitempty"`
}
