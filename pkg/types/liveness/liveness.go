package liveness

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// NewLivenessData creates liveness Data
func NewLivenessData() Result {
	r := &proxy{
		underlying: underlying{
			_once: sync.OnceValue(func() time.Time {
				return time.Now()
			}),
		},
	}
	r.Start()
	return r
}

type Result interface {
	// Start starts probing endpoint liveness
	Start()
	// End ends probing with reason if failed or other detail data
	End(any)
	// RTT reports rtt of liveness probing
	RTT() time.Duration
	// FailureReason denotes failure reason
	FailureReason() error
	// Detail returns check detail
	Detail() any
}

type underlying struct {
	_once     func() time.Time
	timestamp time.Time
	rtt       time.Duration
	reason    error
	detail    any
}

func (d *underlying) Start() {
	d.timestamp = d._once()
}

func (d *underlying) End(v any) {
	d.rtt = time.Since(d.timestamp)
	switch x := v.(type) {
	case error:
		d.reason = x
	default:
		d.detail = x
	}
}

func (d *underlying) RTT() time.Duration {
	return d.rtt
}

func (d *underlying) FailureReason() error {
	return d.reason
}

func (d *underlying) Detail() any {
	return d.detail
}

type proxy struct {
	underlying
}

func (r proxy) MarshalJSON() ([]byte, error) {
	msg := "success"
	if err := r.FailureReason(); err != nil {
		msg = r.FailureReason().Error()
	}

	var d = struct {
		Reachable bool   `json:"reachable"`
		RTT       int64  `json:"rtt(ms)"`
		Message   string `json:"msg,omitempty"`
		Detail    any    `json:"detail,omitempty"`
	}{
		Reachable: r.reason == nil,
		RTT:       r.RTT().Milliseconds(),
		Message:   msg,
		Detail:    r.Detail(),
	}

	return json.Marshal(d)
}

// Checker check remote endpoint liveness
// default Endpoint already implements it using tcp dialer, components should
// override that
type Checker interface {
	LivenessCheck(ctx context.Context) Result
}

type SchemeEndpoint interface {
	Scheme() string
	Key() string
}

func CheckLiveness(ctx context.Context, endpoints ...SchemeEndpoint) map[string]map[string]Result {
	m := map[string]map[string]Result{}
	for _, ep := range endpoints {
		if checker, ok := ep.(Checker); ok {
			if m[ep.Scheme()] == nil {
				m[ep.Scheme()] = map[string]Result{}
			}
			m[ep.Scheme()][ep.Key()] = checker.LivenessCheck(ctx)
		}
	}
	return m
}
