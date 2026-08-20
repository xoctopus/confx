package confws

import (
	"context"
	"net/http"
	"time"

	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/textx"

	"github.com/xoctopus/confx/pkg/conftls"
	"github.com/xoctopus/confx/pkg/types"
)

// Option Endpoint 静态配置与默认钩子.
type Option struct {
	// ListenAddr 监听地址,例如 0.0.0.0:10165.
	ListenAddr string `url:",default=80"`
	// Path HTTP upgrade 路径.
	Path string `url:",default=/"`

	// HandshakeTimeout 握手超时.
	HandshakeTimeout types.Duration `url:",default=5s"`
	// WriteTimeout 单次写超时.
	WriteTimeout types.Duration `url:",default=5s"`
	// IdleTimeout 空闲回收超时;<=0 禁用(SetDefault 后默认为 5m).
	IdleTimeout types.Duration `url:",default=5m"`

	// CheckOriginAllowAll 为 true 时升级握手放行任意 Origin.
	CheckOriginAllowAll bool `url:",default=true"`
	// MaxMessageSize 单帧/消息最大字节.
	MaxMessageSize int `url:",default=32768"`
	// MaxConnection 最大同时在线连接数.
	MaxConnection int `url:",default=65536"`

	// Cert TLS 证书;零值表示明文.
	Cert conftls.X509KeyPair

	// onConnected 见 ConnectionHandler.
	onConnected ConnectionHandler
	// onEstablished 见 EstablishHandler.
	onEstablished EstablishHandler
	// onReceived 非 nil 时进入 Push 模式(框架读循环回调);与 Client.Read 互斥.
	onReceived MessageHandler
}

// SetDefault 按 url tag 填充零值字段.
func (o *Option) SetDefault() {
	must.NoErrorV(textx.SetDefault(o))
}

// ValidateHooks 至少配置 Message(Push) 或 Establish(可在其中自管 Pull).
func (o *Option) ValidateHooks() bool {
	return o.onReceived != nil || o.onEstablished != nil
}

// SetConnectionHandler 设置默认 Connection 钩子.
func (o *Option) SetConnectionHandler(h ConnectionHandler) {
	o.onConnected = h
}

// SetEstablishHandler 设置默认 Establish 钩子.
func (o *Option) SetEstablishHandler(h EstablishHandler) {
	o.onEstablished = h
}

// SetMessageHandler 设置默认 Message 钩子(Push 模式).
func (o *Option) SetMessageHandler(h MessageHandler) {
	o.onReceived = h
}

type (
	// ConnectionHandler 在 Upgrade 之前调用(尚无 Client).
	//
	// 推荐: 读 Header/Query 做门禁或粗鉴权、可回写会话信息等
	// 按需返回本连接 ClientOptionApplier;细鉴权/欢迎帧放到 Establish.
	//
	// 返回 error:框架不 Upgrade,以 HTTP 401 写回 err 文案并结束本次请求.
	// 返回的 ctx 为 nil 时沿用入参 ctx.
	ConnectionHandler func(ctx context.Context, r *http.Request) (context.Context, []ClientOptionApplier, error)

	// EstablishHandler 在 Client 已登记之后、消息循环之前调用(已有稳定 WS).
	//
	// 推荐: 业务握手(二次校验、欢迎帧、订阅/拉模式启动)、可回写上下文(如鉴权结果等)
	// 供后续 Message / Pull 使用;需要拒连时先经 cli 下发业务错误帧再返回 error.
	//
	// 返回 error:框架关闭该 Client 并摘表,不进入消息循环.
	// 返回的 ctx 为 nil 时沿用入参 ctx.
	EstablishHandler func(ctx context.Context, cli Client) (context.Context, error)

	// MessageHandler Push 模式收消息回调;业务需主动 Close 时自行调用 cli.Close.
	MessageHandler func(ctx context.Context, cli Client, t int, data []byte)

	// ReadFailureHandler 读失败回调(cid + 原因);随后连接会关闭.
	ReadFailureHandler func(ctx context.Context, cid string, err error)

	// WriteFailureHandler 写失败回调(cid + 原因);随后连接会关闭.
	WriteFailureHandler func(ctx context.Context, cid string, err error)

	// DisconnectionHandler 连接断开回调(仅触发一次;无 Client,避免误 Close).
	DisconnectionHandler func(ctx context.Context, cid string, err error)
)

// ClientOptionApplier 覆盖单连接 clientOption(由 ConnectionHandler 返回).
type ClientOptionApplier func(*clientOption)

// WithClientMessageHandler 覆盖本连接 Message 钩子.
func WithClientMessageHandler(h MessageHandler) ClientOptionApplier {
	return func(o *clientOption) { o.onReceived = h }
}

// WithClientEstablishHandler 覆盖本连接 Establish 钩子.
func WithClientEstablishHandler(h EstablishHandler) ClientOptionApplier {
	return func(o *clientOption) { o.onEstablished = h }
}

// WithClientReadFailureHandler 设置本连接读失败钩子.
func WithClientReadFailureHandler(h ReadFailureHandler) ClientOptionApplier {
	return func(o *clientOption) { o.onReadFailed = h }
}

// WithClientWriteFailureHandler 设置本连接写失败钩子.
func WithClientWriteFailureHandler(h WriteFailureHandler) ClientOptionApplier {
	return func(o *clientOption) { o.onWriteFailed = h }
}

// WithClientDisconnectionHandler 设置本连接断开钩子.
func WithClientDisconnectionHandler(h DisconnectionHandler) ClientOptionApplier {
	return func(o *clientOption) { o.onDisconnected = h }
}

// WithIdleTimeout 覆盖本连接空闲超时;<=0 禁用.
func WithIdleTimeout(d time.Duration) ClientOptionApplier {
	return func(o *clientOption) { o.idleTimeout = d }
}

// WithWriteTimeout 覆盖本连接写超时.
func WithWriteTimeout(d time.Duration) ClientOptionApplier {
	return func(o *clientOption) { o.writeTimeout = d }
}

type clientOption struct {
	onEstablished  EstablishHandler
	onReceived     MessageHandler
	onReadFailed   ReadFailureHandler
	onWriteFailed  WriteFailureHandler
	onDisconnected DisconnectionHandler
	idleTimeout    time.Duration
	writeTimeout   time.Duration
	underlying     *http.Request

	_detach func(Client)
}

func (o *clientOption) Underlying() *http.Request {
	return o.underlying
}
