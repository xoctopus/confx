package confws

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xoctopus/logx"
	"github.com/xoctopus/x/codex"
)

// Endpoint WebSocket 服务入口:配置、监听、连接生命周期.
//
// 生命周期: Init(资源) → Run(监听) → upgrade(...) → Close.
type Endpoint struct {
	Option

	server   *http.Server
	serveErr atomic.Value // error
	clients  ClientManager
	cancel   context.CancelFunc
	mu       sync.Mutex
	inited   atomic.Bool
	running  atomic.Bool
}

// SetDefault 填充 Option 零值字段.
func (e *Endpoint) SetDefault() {
	e.Option.SetDefault()
}

// Init 初始化资源(TLS、ClientManager);不开始监听.不可重复调用.
func (e *Endpoint) Init(ctx context.Context) (err error) {
	log := logx.From(ctx)
	defer func() {
		if err != nil {
			log.Error(err)
		}
	}()

	if e.inited.Load() {
		return codex.New(ERROR__ENDPOINT_INITIALIZED)
	}

	if !e.Cert.IsZero() {
		if err = e.Cert.Init(); err != nil {
			return fmt.Errorf("failed to init tls cert: %w", err)
		}
	}

	e.clients = newClientManager(e.MaxConnection)
	e.inited.Store(true)
	return nil
}

// Run 开始监听;ctx 作为连接 BaseContext.成功返回即表示已在接受连接.不可重复调用.
func (e *Endpoint) Run(ctx context.Context) (err error) {
	log := logx.From(ctx)
	defer func() {
		log = log.With("addr", e.ListenAddr)
		if err != nil {
			if e.cancel != nil {
				e.cancel()
			}
			log.Error(err)
		} else {
			log.Info("start listening")
		}
	}()

	if !e.inited.Load() {
		return codex.New(ERROR__ENDPOINT_NOT_INITIALIZED)
	}
	if e.running.Load() {
		return codex.New(ERROR__ENDPOINT_INITIALIZED)
	}
	if !e.ValidateHooks() {
		return codex.New(ERROR__MISSING_REQUIRED_HOOKS)
	}

	var root context.Context
	root, e.cancel = context.WithCancel(ctx)

	if err = e.serve(root); err != nil {
		return err
	}
	return nil
}

// Close 取消服务 ctx、关闭全部 Client 并 Shutdown HTTP Server.
func (e *Endpoint) Close(ctx context.Context) error {
	if e.cancel != nil {
		e.cancel()
	}
	if e.clients != nil {
		_ = e.clients.Close(ctx)
	}

	e.mu.Lock()
	srv := e.server
	e.server = nil
	e.mu.Unlock()
	if srv == nil {
		e.running.Store(false)
		return nil
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	err := srv.Shutdown(ctx)
	e.running.Store(false)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	if v := e.serveErr.Load(); v != nil {
		return v.(error)
	}
	return nil
}

func (e *Endpoint) serve(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc(e.Path, e.upgrade)

	srv := &http.Server{
		Addr:    e.ListenAddr,
		Handler: mux,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	ln, err := net.Listen("tcp", e.ListenAddr)
	if err != nil {
		return codex.Wrap(ERROR__FAILED_TO_LISTENING, err)
	}
	e.ListenAddr = ln.Addr().String()

	if !e.Cert.IsZero() {
		srv.TLSConfig = e.Cert.Config()
		ln = tls.NewListener(ln, srv.TLSConfig)
	}

	e.mu.Lock()
	e.server = srv
	e.mu.Unlock()

	e.running.Store(true)
	go func() {
		err := srv.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.serveErr.Store(err)
		}
		e.running.Store(false)
	}()

	return nil
}

func (e *Endpoint) upgrade(w http.ResponseWriter, r *http.Request) {
	var (
		ctx      = r.Context()
		log      = logx.From(ctx)
		err      error
		appliers []ClientOptionApplier
	)

	log.Info("new connection")

	defer func() {
		if err != nil {
			log.Error(err)
		}
	}()

	if e.clients.Activities() >= e.MaxConnection {
		err = codex.New(ERROR__TOO_MANY_CONNECTION)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Connection: 稳定 WS 之前;失败以 HTTP 响应,不 Upgrade.
	if e.onConnected != nil {
		var next context.Context
		next, appliers, err = e.onConnected(ctx, r)
		if err != nil {
			err = codex.Wrap(ERROR__CONNECTION_CALLBACK_FAILED, err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if next != nil {
			ctx = next
		}
	}

	ur, err := (&websocket.Upgrader{
		HandshakeTimeout: time.Duration(e.HandshakeTimeout),
		CheckOrigin:      func(*http.Request) bool { return e.CheckOriginAllowAll },
	}).Upgrade(w, r, nil)
	if err != nil {
		err = codex.Wrap(ERROR__FAILED_TO_UPGRADE, err)
		return
	}
	ur.SetReadLimit(int64(e.MaxMessageSize))

	opt := e.option(appliers...)
	opt.underlying = r
	cli := newClient(ctx, ur, opt)

	defer func() {
		_ = cli.Close(ctx)
	}()

	if err = e.clients.Add(ctx, cli); err != nil {
		err = codex.Wrap(ERROR__FAILED_TO_REGISTER_CLIENT, err)
		return
	}

	// Establish: 稳定 WS 之后、消息循环前;可继续改写 ctx.
	if opt.onEstablished != nil {
		var next context.Context
		next, err = opt.onEstablished(ctx, cli)
		if err != nil {
			_ = cli.Close(ctx)
			return
		}
		if next != nil {
			ctx = next
		}
	}

	cli.start(ctx)

	select {
	case <-ctx.Done():
		err = codex.Wrap(ERROR__SERVER_CLOSED, cli.Close(ctx))
	case <-cli.Done():
		err = codex.Wrap(ERROR__CLIENT_CLOSED, cli.Err())
	}
}

// option 生成客户端连接选项
func (e *Endpoint) option(appliers ...ClientOptionApplier) *clientOption {
	o := &clientOption{
		onEstablished: e.onEstablished,
		onReceived:    e.onReceived,
		idleTimeout:   time.Duration(e.IdleTimeout),
		writeTimeout:  time.Duration(e.WriteTimeout),
	}

	if e.clients != nil {
		o._detach = e.clients.Detach
	}

	for _, f := range appliers {
		f(o)
	}

	return o
}

// Client 按 id 取在线连接;不存在或未 Init 返回 nil.
func (e *Endpoint) Client(id string) Client {
	if e.clients == nil {
		return nil
	}
	return e.clients.Get(id)
}

// Session 实现 Session 接口;等价于 Client(id).
func (e *Endpoint) Session(id string) Client {
	return e.Client(id)
}

// WithContext 注入 Session,供业务按连接 ULID 查找 / 关闭 Client.
func (e *Endpoint) WithContext(ctx context.Context) context.Context {
	return WithSession(ctx, e)
}

// String 返回可拨号的 WebSocket URL(ws/wss://ListenAddr+Path);Cert 有效时为 wss.
func (e *Endpoint) String() string {
	scheme := "ws"
	if !e.Cert.IsZero() {
		scheme = "wss"
	}
	return (&url.URL{Scheme: scheme, Host: e.ListenAddr, Path: e.Path}).String()
}
