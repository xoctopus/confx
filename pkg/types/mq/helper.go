package mq

import (
	"context"
	"encoding"
	"encoding/json"
	"hash/crc32"
	"hash/fnv"
	"math"
)

func MarshalV(v any) ([]byte, error) {
	switch x := v.(type) {
	case nil:
		return nil, nil
	case []byte:
		return x, nil
	case string:
		return []byte(x), nil
	case encoding.TextMarshaler:
		return x.MarshalText()
	case json.Marshaler:
		return x.MarshalJSON()
	default:
		return json.Marshal(x)
	}
}

func UnmarshalV(data []byte, v any) error {
	if v == nil && len(data) == 0 {
		return nil
	}
	switch x := v.(type) {
	case *[]byte:
		*x = data
		return nil
	case *string:
		*x = string(data)
		return nil
	case encoding.TextUnmarshaler:
		return x.UnmarshalText(data)
	case json.Unmarshaler:
		return x.UnmarshalJSON(data)
	default:
		return json.Unmarshal(data, x)
	}
}

// Hasher help to hash message.Key to dispatch message to task worker
type Hasher func(string) uint16

func Fnv(k string) uint16 {
	h := fnv.New32a()
	h.Reset()
	_, _ = h.Write([]byte(k))
	return uint16(h.Sum32() % math.MaxUint16)
}

func CRC(k string) uint16 {
	return uint16(crc32.ChecksumIEEE([]byte(k)) & math.MaxUint16)
}

type ConsumeMode int

const (
	// GlobalOrdered messages are processed strictly one by one in the order
	// they were received globally.
	GlobalOrdered ConsumeMode = iota
	// PartitionOrdered messages with the same partition key are processed sequentially,
	// while messages with different keys are handled in parallel.
	PartitionOrdered
	// Concurrent messages are processed in parallel with no guarantee of ordering.
	// this mode offers the highest throughput.
	Concurrent
)

type Handler func(context.Context, Message) error
