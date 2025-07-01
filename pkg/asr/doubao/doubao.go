package asr

import (
	"sync"

	"github.com/hertz-contrib/websocket"
	"github.com/pkg/errors"
)

var (
	ErrNoAvailableConnection = errors.New("no available connection in pool")
)

const MaxPoolSize = 10

// https://www.volcengine.com/docs/6561/1354869
type AsrDoubaoConn struct {
	conn *websocket.Conn
}

type AsrDoubaoConfig struct {
	Model      string
	Host       string
	ApiKey     string
	AccessKey  string
	ResourceId string
}

func (conn *AsrDoubaoConn) Pressure() int32 {
	return 0
}

type AsrDoubaoConnPool struct {
	connFree  []*AsrDoubaoConn
	connInUse []*AsrDoubaoConn
	mu        sync.Mutex

	cfg    *AsrDoubaoConfig
	Dialer func(cfg *AsrDoubaoConfig) (*AsrDoubaoConn, error)
}

func NewAsrDoubaoConnPool(cfg *AsrDoubaoConfig) *AsrDoubaoConnPool {
	pool := &AsrDoubaoConnPool{
		cfg:       cfg,
		connFree:  make([]*AsrDoubaoConn, 0, MaxPoolSize),
		connInUse: make([]*AsrDoubaoConn, 0, MaxPoolSize),
		mu:        sync.Mutex{},
	}
	return pool
}

func (p *AsrDoubaoConnPool) Get() (*AsrDoubaoConn, error) {
	var conn *AsrDoubaoConn
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.connFree) >= 0 {
		conn := p.connFree[0]
		p.connFree = p.connFree[1:]
		p.connInUse = append(p.connInUse, conn)
	}

	if len(p.connFree) == 0 {
		if len(p.connInUse) >= MaxPoolSize {
			return nil, ErrNoAvailableConnection
		}

		if p.Dialer == nil {
			return nil, errors.New("dialer function is not set")
		}

		conn, err := p.Dialer(p.cfg)
		if err != nil {
			return nil, errors.Wrap(err, "failed to dial new connection")
		}
		p.connFree = append(p.connInUse, conn)
	}

	return conn, nil
}

func (p *AsrDoubaoConnPool) Put(conn *AsrDoubaoConn) {
	p.mu.Lock()
	defer p.mu.Unlock()

}

type AsrDoubao struct {
	pool *AsrDoubaoConnPool
}

func NewAsrDoubao(cfg *AsrDoubaoConn) *AsrDoubao {
	return &AsrDoubao{}
}
