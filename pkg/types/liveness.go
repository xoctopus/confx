package types

import (
	"context"
)

// LivenessData presents LivenessChecker result
type LivenessData struct {
	// Reachable if remote endpoint is reachable
	Reachable bool `json:"reachable"`
	// TTL probes ttl to remote endpoint
	TTL Duration `json:"ttl,omitempty"`
	// Message result or extended message
	Message string `json:"msg,omitempty"`
}

// LivenessChecker check remote endpoint liveness
// Endpoint already implements it, components should override that
type LivenessChecker interface {
	LivenessCheck(ctx context.Context) LivenessData
}

type SchemedEndpoint interface {
	Scheme() string
	Endpoint() string
}

func CheckLiveness(ctx context.Context, endpoints ...SchemedEndpoint) map[string]map[string]LivenessData {
	m := map[string]map[string]LivenessData{}
	for _, ep := range endpoints {
		if x, ok := ep.(LivenessChecker); ok {
			if m[ep.Scheme()] == nil {
				m[ep.Scheme()] = map[string]LivenessData{}
			}
			m[ep.Scheme()][ep.Endpoint()] = x.LivenessCheck(ctx)
		}
	}
	return m
}
