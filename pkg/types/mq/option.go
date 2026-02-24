package mq

import (
	"context"
	"hash/crc32"
	"hash/fnv"
	"math"
)

type Option interface {
	OptionScheme() string
}

type OptionApplier interface {
	Apply(Option)
}

type OptionApplyFunc func(Option)

func (f OptionApplyFunc) Apply(opt Option) { f(opt) }

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

type ConsumeHandleMode int

const (
	// GlobalOrdered messages are processed strictly one by one in the order
	// they were received globally.
	GlobalOrdered ConsumeHandleMode = iota
	// PartitionOrdered messages with the same partition key are processed sequentially,
	// while messages with different keys are handled in parallel.
	PartitionOrdered
	// Concurrent messages are processed in parallel with no guarantee of ordering.
	// this mode offers the highest throughput.
	Concurrent
)

type (
	AsyncPubCallback[PM any] func(PM, error)
	SubHandler[CM any]       func(context.Context, CM) error
	SubCallback[CM any]      func(Acknowledger[CM], CM, error)
)
