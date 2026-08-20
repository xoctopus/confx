package hack

import (
	"context"
	"testing"

	"github.com/gorilla/websocket"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/confx/pkg/confws"
	"github.com/xoctopus/confx/pkg/types"
)

// NewWSEndpoint 初始化并监听可拨号的 confws.Endpoint(127.0.0.1:0,/ws,默认禁用空闲回收).
// setup 在 Init 前调用,用于挂钩子/改配置;测试结束自动 Close.
func NewWSEndpoint(t testing.TB, setup func(*confws.Endpoint)) *confws.Endpoint {
	t.Helper()
	ep := &confws.Endpoint{}
	ep.SetDefault()
	ep.ListenAddr = "127.0.0.1:0"
	ep.Path = "/ws"
	ep.IdleTimeout = types.Duration(-1)
	if setup != nil {
		setup(ep)
	}
	Expect(t, ep.Init(context.Background()), Succeed())
	Expect(t, ep.Run(context.Background()), Succeed())
	t.Cleanup(func() { _ = ep.Close(context.Background()) })
	return ep
}

// DialWS 拨到 ep.String();失败则 Fatal;测试结束自动 Close.
func DialWS(t testing.TB, ep *confws.Endpoint) *websocket.Conn {
	t.Helper()
	c, _, err := websocket.DefaultDialer.Dial(ep.String(), nil)
	Expect(t, err, Succeed())
	t.Cleanup(func() { _ = c.Close() })
	return c
}
