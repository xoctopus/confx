package liveness

import (
	"context"
	"fmt"
	"time"
)

// NewLivenessData creates liveness Data
func NewLivenessData() Result {
	r := &proxy{}
	r.Start()
	return r
}

type Result interface {
	// Start starts probing endpoint liveness
	Start()
	// End ends probing with reason if failed
	End(error)
	// RTT reports rtt of liveness probing
	RTT() time.Duration
	// FailureReason denotes failure reason
	FailureReason() error
}

type underlying struct {
	timestamp time.Time
	rtt       time.Duration
	reason    error
}

func (d *underlying) Start() {
	d.timestamp = time.Now()
}

func (d *underlying) End(err error) {
	d.rtt = time.Since(d.timestamp)
	d.reason = err
}

func (d *underlying) RTT() time.Duration {
	return d.rtt
}

func (d *underlying) FailureReason() error {
	return d.reason
}

type proxy struct {
	underlying
}

func (r proxy) MarshalJSON() ([]byte, error) {
	if r.reason == nil {
		return fmt.Appendf(nil,
			`{"reachable":true,"rtt(ms)":%d,"msg":"success"}`,
			r.RTT().Milliseconds(),
		), nil
	}

	return fmt.Appendf(nil,
		`{"reachable":false,"rtt(ms)":%d,"msg":"%s"}`,
		r.RTT().Milliseconds(), r.FailureReason().Error(),
	), nil
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
