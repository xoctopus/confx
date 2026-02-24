package confpulsar

// ExtraKind defines types of extended metadata for pulsar
//
// for extending more mq-specific features, implement mq.HasExtra, mq.CanAppendExtra
// +genx:enum
type ExtraKind uint8

const (
	EXTRA_KIND_UNKNOWN ExtraKind = iota
	EXTRA_KIND__LAST_SEQUENCE_ID
	EXTRA_KIND__DELIVERY_DELAYED
)
