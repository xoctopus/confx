package confkafka

// this file defines keys for extended metadata and more mq-specific features for kafka

const (
	// EXTRA_KEY__EXPIRED_AT is the header key for message expiration timestamp (epoch seconds).
	EXTRA_KEY__EXPIRED_AT = "EXPIRED_AT"
	// EXTRA_KEY__RETRY_COUNT is the header key for redelivery/retry count
	EXTRA_KEY__RETRY_COUNT = "RETRY_COUNT"
)
