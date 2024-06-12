package confmqtt

import (
	"bytes"

	"github.com/pkg/errors"
)

type QOS int8

const (
	QOS_UNKNOWN        QOS = iota - 1
	QOS__ONCE              // 0
	QOS__AT_LEAST_ONCE     // 1
	QOS__ONLY_ONCE         // 2
)

var InvalidQOS = errors.New("invalid QOS type")

func ParseQOSFromString(s string) (QOS, error) {
	switch s {
	default:
		return QOS_UNKNOWN, InvalidQOS
	case "":
		return QOS_UNKNOWN, nil
	case "ONCE":
		return QOS__ONCE, nil
	case "AT_LEAST_ONCE":
		return QOS__AT_LEAST_ONCE, nil
	case "ONLY_ONCE":
		return QOS__ONLY_ONCE, nil
	}
}

func ParseQOSFromLabel(s string) (QOS, error) {
	switch s {
	default:
		return QOS_UNKNOWN, InvalidQOS
	case "":
		return QOS_UNKNOWN, nil
	case "0":
		return QOS__ONCE, nil
	case "1":
		return QOS__AT_LEAST_ONCE, nil
	case "2":
		return QOS__ONLY_ONCE, nil
	}
}

func (v QOS) Int() int {
	return int(v)
}

func (v QOS) String() string {
	switch v {
	default:
		return "UNKNOWN"
	case QOS_UNKNOWN:
		return ""
	case QOS__ONCE:
		return "ONCE"
	case QOS__AT_LEAST_ONCE:
		return "AT_LEAST_ONCE"
	case QOS__ONLY_ONCE:
		return "ONLY_ONCE"
	}
}

func (v QOS) Label() string {
	switch v {
	default:
		return "UNKNOWN"
	case QOS_UNKNOWN:
		return ""
	case QOS__ONCE:
		return "0"
	case QOS__AT_LEAST_ONCE:
		return "1"
	case QOS__ONLY_ONCE:
		return "2"
	}
}

func (v QOS) MarshalText() ([]byte, error) {
	s := v.String()
	if s == "UNKNOWN" {
		return nil, InvalidQOS
	}
	return []byte(s), nil
}

func (v *QOS) UnmarshalText(data []byte) error {
	s := string(bytes.ToUpper(data))
	val, err := ParseQOSFromString(s)
	if err != nil {
		return err
	}
	*v = val
	return nil
}
