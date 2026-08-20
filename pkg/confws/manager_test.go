package confws

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/xoctopus/x/testx"
)

type stubClient struct {
	id     string
	closed atomic.Bool
	done   chan struct{}
	once   sync.Once
	closes atomic.Int32
}

func newStub(id string) *stubClient {
	return &stubClient{id: id, done: make(chan struct{})}
}

func (s *stubClient) Close(context.Context) error {
	s.closes.Add(1)
	s.closed.Store(true)
	s.once.Do(func() { close(s.done) })
	return nil
}
func (s *stubClient) Done() <-chan struct{} { return s.done }
func (s *stubClient) Err() error            { return nil }
func (s *stubClient) Read(context.Context) (int, []byte, error) {
	return 0, nil, nil
}
func (s *stubClient) WriteBinary(context.Context, []byte) error { return nil }
func (s *stubClient) WriteText(context.Context, []byte) error   { return nil }
func (s *stubClient) LastActivity() time.Time                   { return time.Time{} }
func (s *stubClient) ID() string                                { return s.id }
func (s *stubClient) Underlying() *http.Request                 { return nil }

func TestManager_AddGetDetachRemove(t *testing.T) {
	ctx := context.Background()
	m := newClientManager(2)

	a := newStub("a")
	Expect(t, m.Add(ctx, a), Succeed())
	Expect(t, m.Activities(), Equal(1))
	Expect(t, m.Get("a"), Be[Client](a))
	Expect(t, m.Get("missing"), BeNil[Client]())

	m.Detach(a)
	Expect(t, m.Activities(), Equal(0))
	Expect(t, m.Get("a"), BeNil[Client]())
	Expect(t, a.closes.Load(), Equal(int32(0)))

	Expect(t, m.Add(ctx, a), Succeed())
	Expect(t, m.Remove(ctx, "a"), Succeed())
	Expect(t, a.closes.Load(), Equal(int32(1)))
	// stub Close 不走 _detach; 显式 Detach 模拟真实 Client
	m.Detach(a)
	Expect(t, m.Activities(), Equal(0))
}

func TestManager_SwapClosesOld(t *testing.T) {
	ctx := context.Background()
	m := newClientManager(2)

	old := newStub("same")
	newer := newStub("same")
	Expect(t, m.Add(ctx, old), Succeed())
	Expect(t, m.Add(ctx, newer), Succeed())
	Expect(t, old.closes.Load(), Equal(int32(1)))
	Expect(t, m.Get("same"), Be[Client](newer))
	Expect(t, m.Activities(), Equal(1))
}

func TestManager_DetachCompareAndDelete(t *testing.T) {
	ctx := context.Background()
	m := newClientManager(2)

	old := newStub("same")
	newer := newStub("same")
	Expect(t, m.Add(ctx, old), Succeed())
	Expect(t, m.Add(ctx, newer), Succeed())

	m.Detach(old) // 不应删掉 newer
	Expect(t, m.Get("same"), Be[Client](newer))
	Expect(t, m.Activities(), Equal(1))

	m.Detach(newer)
	Expect(t, m.Activities(), Equal(0))
}

func TestManager_Threshold(t *testing.T) {
	ctx := context.Background()
	m := newClientManager(1)

	Expect(t, m.Add(ctx, newStub("a")), Succeed())
	err := m.Add(ctx, newStub("b"))
	Expect(t, err, Failed())
	Expect(t, err.Error(), ContainsSubString("reached max connection"))

	// 同 id 替换不受 threshold 阻挡
	Expect(t, m.Add(ctx, newStub("a")), Succeed())
	Expect(t, m.Activities(), Equal(1))
}

func TestManager_CloseAll(t *testing.T) {
	ctx := context.Background()
	m := newClientManager(8)
	a, b := newStub("a"), newStub("b")
	Expect(t, m.Add(ctx, a), Succeed())
	Expect(t, m.Add(ctx, b), Succeed())
	Expect(t, m.Close(ctx), Succeed())
	Expect(t, a.closes.Load(), Equal(int32(1)))
	Expect(t, b.closes.Load(), Equal(int32(1)))
}

func TestManager_AddInvalid(t *testing.T) {
	ctx := context.Background()
	m := newClientManager(1)
	Expect(t, m.Add(ctx, nil), Failed())
	Expect(t, m.Add(ctx, newStub("")), Failed())
	Expect(t, m.Remove(ctx, ""), Failed())
	Expect(t, m.Remove(ctx, "missing"), Succeed())
}
