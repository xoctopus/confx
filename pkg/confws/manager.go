package confws

import (
	"context"
	"fmt"

	"github.com/xoctopus/x/syncx"
)

// ClientManager 在线连接登记表.
//
//	Detach: 仅摘表(不关连接;Client.Close 内部调用)
//	Remove: 关闭指定连接(经 Client.Close → Detach)
//	Close:  关闭全部连接
type ClientManager interface {
	// Add 登记客户端;同 id 已存在则关闭旧连接并替换.
	Add(ctx context.Context, c Client) error
	// Detach 仅当 map 中仍是该实例时摘表.
	Detach(c Client)
	// Remove 关闭指定客户端.
	Remove(ctx context.Context, id string) error
	// Get 返回指定客户端;不存在返回 nil.
	Get(id string) Client
	// Activities 当前登记连接数.
	Activities() int
	// Close 关闭所有登记连接.
	Close(ctx context.Context) error
}

type manager struct {
	clients   syncx.Map[string, Client]
	threshold int
}

var _ ClientManager = (*manager)(nil)

func newClientManager(threshold int) ClientManager {
	if threshold <= 0 {
		threshold = 1
	}
	return &manager{
		threshold: threshold,
		clients:   syncx.NewXmap[string, Client](),
	}
}

func (m *manager) Activities() int {
	return m.clients.Len()
}

func (m *manager) Get(id string) Client {
	if id == "" {
		return nil
	}
	v, ok := m.clients.Load(id)
	if !ok {
		return nil
	}
	return v
}

func (m *manager) Add(ctx context.Context, c Client) error {
	if c == nil {
		return fmt.Errorf("invalid client, got nil")
	}
	if len(c.ID()) == 0 {
		return fmt.Errorf("invalid client id, got empty")
	}

	if _, exists := m.clients.Load(c.ID()); !exists && m.clients.Len() >= m.threshold {
		return fmt.Errorf("reached max connection: %d", m.clients.Len())
	}

	if x, ok := m.clients.Swap(c.ID(), c); ok && x != nil && x != c {
		_ = x.Close(ctx)
	}
	return nil
}

// Detach 仅摘表;CompareAndDelete 避免误删同 id 的新连接.
func (m *manager) Detach(c Client) {
	if c == nil || c.ID() == "" {
		return
	}
	m.clients.CompareAndDelete(c.ID(), c)
}

// Remove 关闭指定连接.
func (m *manager) Remove(ctx context.Context, id string) error {
	if len(id) == 0 {
		return fmt.Errorf("invalid client id, got empty")
	}
	c := m.Get(id)
	if c == nil {
		return nil
	}
	return c.Close(ctx)
}

// Close 关闭全部连接.
func (m *manager) Close(ctx context.Context) error {
	list := make([]Client, 0, m.clients.Len())
	m.clients.Range(func(_ string, c Client) bool {
		if c != nil {
			list = append(list, c)
		}
		return true
	})
	for _, c := range list {
		_ = c.Close(ctx)
	}
	return nil
}
