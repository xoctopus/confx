package types

import (
	"context"
	"net/url"
	"time"
)

type LivenessChecker interface {
	LivenessCheck() map[string]string
}

type LivenessCheckDetail struct {
	Reachable bool          `json:"reachable"`
	TTL       time.Duration `json:"ttl,omitempty"`
	Msg       string        `json:"msg,omitempty"`
}

type ComponentLivenessChecker interface {
	LivenessCheck(ctx context.Context) map[Component]LivenessCheckDetail
}

type Component interface {
	// Key identifies a unique component with
	Key() string
	// Hostname endpoint hostname
	Hostname() string
	// Options with key and values
	Options() url.Values
}
