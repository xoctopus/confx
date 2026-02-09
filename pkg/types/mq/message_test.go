package mq_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/xoctopus/sfid/pkg/sfid"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/types/mq"
)

type JSONArshaler struct {
	X int    `json:"x"`
	Y string `json:"y"`
}

func (x JSONArshaler) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"x":%d,"y":%q}`, x.X, x.Y)), nil
}

func (x *JSONArshaler) UnmarshalJSON(data []byte) error {
	m := map[string]any{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	x.X = int(m["x"].(float64))
	x.Y = m["y"].(string)
	return nil
}

type TextArshaler struct {
	X int    `json:"x"`
	Y string `json:"y"`
}

func (x TextArshaler) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"x":%d,"y":%q}`, x.X, x.Y)), nil
}

func (x *TextArshaler) UnmarshalText(data []byte) error {
	m := map[string]any{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	x.X = int(m["x"].(float64))
	x.Y = m["y"].(string)
	return nil
}

type Data struct {
	X int    `json:"x"`
	Y string `json:"y"`
}

type Arshaler interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

func TestMessage(t *testing.T) {
	ctx := sfid.With(context.Background(), sfid.NewDefaultIDGen(100))

	t.Run("Message", func(t *testing.T) {
		for _, v := range []any{
			&JSONArshaler{X: 1, Y: "2"},
			&TextArshaler{X: 1, Y: "2"},
			&Data{X: 1, Y: "2"},
			JSONArshaler{X: 1, Y: "2"},
			TextArshaler{X: 1, Y: "2"},
			Data{X: 1, Y: "2"},
		} {
			m := mq.NewMessage(ctx, "topic", v)
			Expect(t, m.Topic(), Equal("topic"))
			Expect(t, m.Data(), Equal([]byte(`{"x":1,"y":"2"}`)))
			m.(mq.MutMessage).AddExtra("k1", "v1")
			m.(mq.MutMessage).AddExtra("k1", "v2")
			m.(mq.MutMessage).AddExtra("k2", "v1")
			m.(mq.MutMessage).AddExtra("k2", "v2")

			data, err := m.(Arshaler).Marshal()
			Expect(t, err, Succeed())

			x, err := mq.ParseMessage(data)
			Expect(t, x.Topic(), Equal(m.Topic()))
			Expect(t, x.Data(), Equal(m.Data()))
			Expect(t, x.ID(), Equal(m.ID()))
			Expect(t, x.Timestamp().UnixNano(), Equal(m.Timestamp().UnixNano()))
			Expect(t, x.Extra(), Equal(m.Extra()))
		}
	})

	t.Run("Ordered", func(t *testing.T) {
		id := sfid.Must(ctx).MustID()
		m := mq.NewMessageWithID(t.Name(), id, "1")
		m.(mq.MutMessage).SetPubOrderedKey("p_ordered")
		m.(mq.MutMessage).SetSubOrderedKey("s_ordered")
		Expect(t, m.(mq.OrderedMessage).PubOrderedKey(), Equal("p_ordered"))
		Expect(t, m.(mq.OrderedMessage).SubOrderedKey(), Equal("s_ordered"))
	})

	t.Run("Helper", func(t *testing.T) {
		data, err := mq.MarshalV("123")
		Expect(t, err, Succeed())
		Expect(t, data, Equal([]byte("123")))

		data, err = mq.MarshalV([]byte("123"))
		Expect(t, err, Succeed())
		Expect(t, data, Equal([]byte("123")))

		data, err = mq.MarshalV(nil)
		Expect(t, err, Succeed())
		Expect(t, data, Equal[[]byte](nil))

		x := any(new(string))
		Expect(t, mq.UnmarshalV([]byte("123"), x), Succeed())
		Expect(t, *(x.(*string)), Equal("123"))

		x = any(new([]byte))
		Expect(t, mq.UnmarshalV([]byte("123"), x), Succeed())
		Expect(t, *(x.(*[]byte)), Equal([]byte("123")))

		x = any(new(int))
		Expect(t, mq.UnmarshalV([]byte("123"), x), Succeed())
		Expect(t, *(x.(*int)), Equal(123))

		x = any(new(JSONArshaler))
		Expect(t, mq.UnmarshalV([]byte(`{"x":1,"y":"2"}`), x), Succeed())
		Expect(t, *(x.(*JSONArshaler)), Equal(JSONArshaler{X: 1, Y: "2"}))

		x = any(new(TextArshaler))
		Expect(t, mq.UnmarshalV([]byte(`{"x":1,"y":"2"}`), x), Succeed())
		Expect(t, *(x.(*TextArshaler)), Equal(TextArshaler{X: 1, Y: "2"}))

		Expect(t, mq.UnmarshalV(nil, nil), Succeed())
	})
	t.Run("Failed", func(t *testing.T) {
		_, err := mq.ParseMessage([]byte(`{`))
		Expect(t, err, Failed())
	})
}
