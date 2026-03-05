package confxxl_test

import (
	"context"
	"testing"
	"time"

	. "cgtech.gitlab.com/saitox/x/testx"

	"cgtech.gitlab.com/saitox/confx/hack"
	"cgtech.gitlab.com/saitox/confx/pkg/confxxl"
)

func TestEndpoint(t *testing.T) {
	var (
		ctx    = hack.Context(t)
		cancel context.CancelFunc
	)
	ctx, cancel = context.WithTimeout(ctx, time.Minute)
	defer cancel()
	ctx = hack.WithXXLRegistry(ctx, t, "http://localhost:18081/xxl-job-admin", "confx")

	registry := confxxl.Must(ctx)
	defer func() { _ = registry.Close() }()
	err := registry.RegisterJob(
		"confx",      // executor name
		"confx_test", // job name
		func(ctx context.Context, r *confxxl.TriggerRequest) error {
			t.Log(r.LogID, r.ExecutorHandler, time.Now().Unix())
			return nil
		},
	)
	Expect(t, err, Succeed())

	active := false
	for {
		select {
		case <-ctx.Done():
			Expect(t, active, BeTrue())
			return
		default:
			if registry.IsActive("confx") {
				active = true
			} else {
				active = false
			}
			time.Sleep(time.Second)
		}
	}
}
