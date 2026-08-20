package confws

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/oklog/ulid/v2"
	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/codex"
)

// Client 已升级的 WebSocket 会话.
//
// Push(有 MessageHandler):框架读循环回调,业务勿调用 Read.
// Pull(无 MessageHandler):业务自行 Read;写均经 WriteText/WriteBinary.
type Client interface {
	// Close 关闭连接并从管理器 Detach;可重复调用.
	Close(ctx context.Context) error

	// Done 关闭完成信号.
	Done() <-chan struct{}
	// Err 关闭原因;未关闭时为 nil.
	Err() error

	// Read 拉模式读一帧;Push 模式返回 ERROR__READ_ON_NON_PULL_CLIENT.
	Read(ctx context.Context) (t int, data []byte, err error)
	// WriteBinary 写二进制帧;失败将关闭连接.
	WriteBinary(ctx context.Context, data []byte) error
	// WriteText 写文本帧;失败将关闭连接.
	WriteText(ctx context.Context, data []byte) error

	// LastActivity 最近读写时间.
	LastActivity() time.Time

	// ID 连接标识(ULID,由框架生成).
	ID() string

	// Underlying 返回原始http请求
	Underlying() *http.Request
}

func newClient(ctx context.Context, c *websocket.Conn, opt *clientOption) *client {
	cli := &client{
		clientOption: opt,

		id:   ulid.Make().String(),
		conn: c,
		done: make(chan struct{}),
	}
	cli.log = logx.From(ctx).With("client", cli.id)

	cli.touch()

	return cli
}

type client struct {
	*clientOption

	raw *http.Request
	id  string
	log logx.Logger
	mu  sync.Mutex

	conn   *websocket.Conn
	last   atomic.Int64
	closed atomic.Bool
	done   chan struct{}
	once   sync.Once
	err    atomic.Value
}

func (c *client) touch() {
	c.last.Store(time.Now().UnixNano())
}

func (c *client) ID() string {
	return c.id
}

func (c *client) Read(ctx context.Context) (t int, data []byte, err error) {
	if c.onReceived != nil {
		return 0, nil, codex.New(ERROR__READ_ON_NON_PULL_CLIENT)
	}

	c.touch()
	if c.closed.Load() {
		return 0, nil, codex.New(ERROR__CLIENT_CLOSED)
	}

	return c.conn.ReadMessage()
}

func (c *client) write(ctx context.Context, t int, data []byte) error {
	c.touch()

	if c.closed.Load() {
		return codex.New(ERROR__CLIENT_CLOSED)
	}
	// TODO - PERF 不支持写保护 为避免写静态 写消息在 mu 临界区内. 未来考虑支持异步写+回调
	c.mu.Lock()
	if c.writeTimeout > 0 {
		_ = c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	}
	err := c.conn.WriteMessage(t, data)
	c.mu.Unlock()
	if err != nil {
		if c.onWriteFailed != nil {
			c.onWriteFailed(ctx, c.id, err)
		}
		_ = c.close(ctx, err)
		return err
	}
	return nil
}

func (c *client) WriteBinary(ctx context.Context, data []byte) error {
	return c.write(ctx, websocket.BinaryMessage, data)
}

func (c *client) WriteText(ctx context.Context, data []byte) error {
	return c.write(ctx, websocket.TextMessage, data)
}

func (c *client) Close(ctx context.Context) error {
	return c.close(ctx, codex.New(ERROR__CLIENT_CLOSED))
}

// close 记录关闭原因并关闭连接;首次调用生效,并触发 onDisconnected 一次.
func (c *client) close(ctx context.Context, reason error) error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}
	if reason == nil {
		reason = codex.New(ERROR__CLIENT_CLOSED)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	c.log.Error(fmt.Errorf("client closed caused by: %v", reason))

	c.err.Store(reason)

	err := c.conn.Close()
	if c._detach != nil {
		c._detach(c)
	}
	c.once.Do(func() { close(c.done) })

	if c.onDisconnected != nil {
		c.onDisconnected(ctx, c.id, reason)
	}
	return err
}

func (c *client) LastActivity() time.Time {
	return time.Unix(0, c.last.Load())
}

func (c *client) Done() <-chan struct{} {
	return c.done
}

func (c *client) Err() error {
	if v := c.err.Load(); v != nil {
		return v.(error)
	}
	return nil
}

func (c *client) loop(ctx context.Context) {
	if c.onReceived == nil {
		return
	}

	go func() {
		select {
		case <-ctx.Done():
			_ = c.close(ctx, ctx.Err())
		case <-c.done:
		}
	}()

	go func() {
		defer func() {
			_ = c.close(ctx, codex.New(ERROR__CLIENT_CLOSED))
		}()

		for {
			if c.closed.Load() {
				return
			}

			t, data, err := c.conn.ReadMessage()
			if err != nil {
				if c.closed.Load() {
					return
				}
				if c.onReadFailed != nil {
					c.onReadFailed(ctx, c.id, err)
				}
				_ = c.close(ctx, err)
				return
			}

			c.touch()
			// TODO - PERF hook panic during onReceived
			c.onReceived(ctx, c, t, data)
		}
	}()
}

func (c *client) idle(ctx context.Context) {
	if c.idleTimeout <= 0 {
		return
	}
	interval := max(c.idleTimeout/2, time.Second)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-c.done:
				return
			case <-ticker.C:
				if time.Since(c.LastActivity()) >= c.idleTimeout {
					_ = c.close(ctx, codex.New(ERROR__CLIENT_IDLE_TIMEOUT))
					return
				}
			}
		}
	}()
}

func (c *client) start(ctx context.Context) {
	c.loop(ctx)
	c.idle(ctx)
}
