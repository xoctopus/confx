package types

import (
	"context"
)

type LivenessData struct {
	Reachable bool     `json:"reachable"`
	TTL       Duration `json:"ttl,omitempty"`
	Message   string   `json:"msg,omitempty"`
}

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
