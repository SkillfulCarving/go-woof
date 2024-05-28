package server

import (
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type ProxyConn struct {
	ID       uint32
	Conn     net.Conn
	TimeUnix time.Time
}

type ProxyPool struct {
	Seq  *Sequencer
	data *sync.Map
}

// NewProxyPool 创建新的代理池
func NewProxyPool() *ProxyPool {
	return &ProxyPool{
		Seq:  NewSequencer(),
		data: &sync.Map{},
	}
}

func (p *ProxyPool) CreateProxyConn(conn net.Conn) *ProxyConn {
	proxyConn := &ProxyConn{
		ID:       p.Seq.Next(),
		Conn:     conn,
		TimeUnix: time.Now(),
	}
	if proxyConn.ID == 0 {
		proxyConn.ID = p.Seq.Next()
	}
	p.data.Store(proxyConn.ID, proxyConn)
	return proxyConn
}

func (p *ProxyPool) GetProxyConn(id uint32) *ProxyConn {
	if pool, ok := p.data.Load(id); ok {
		return pool.(*ProxyConn)
	}
	return &ProxyConn{}
}

func (p *ProxyPool) CloseProxyConn(id uint32) {
	if val, ok := p.data.LoadAndDelete(id); ok {
		pool := val.(*ProxyConn)
		_ = pool.Conn.Close()
	}
}

func (p *ProxyPool) ProxyPoolClose() {
	p.data.Range(func(key, value any) bool {
		p.data.Delete(key)
		_ = value.(*ProxyConn).Conn.Close()
		return true
	})
}

type Sequencer struct {
	current uint32
}

// NewSequencer 创建TCP序列号
func NewSequencer() *Sequencer {
	return &Sequencer{current: math.MaxUint32}
}

// Next 返回下一个TCP序列号
func (t *Sequencer) Next() uint32 {
	value := atomic.AddUint32(&t.current, 1) // 增加
	return value
}
