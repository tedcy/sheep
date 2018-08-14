package simple_conn_pool

import (
	"container/list"
	"sync"
	"net"
	"fmt"
	"time"
)

const minMinConn = 5

type pConn struct {
	net.Conn
	owner *ConnPool
}

func newPConn(conn net.Conn, owner *ConnPool) net.Conn {
	return &pConn{
		Conn:	conn,
		owner:	owner,
	}
}

func (this *pConn) Close() error{
	this.owner.put(this)
	return nil
}

type ConnPoolConfig struct {
	Addr				string
	MaxConn				int
	MinConn				int
	ConnectTimeout		time.Duration	
}

type ConnPool struct {
	lock			sync.Mutex
	conns			list.List
	config			*ConnPoolConfig
	count			int
}

func NewConnPool(config *ConnPoolConfig) (pool *ConnPool, err error) {
	if (config == nil) {
		return nil, fmt.Errorf("invalid config: nil ptr")
	}
	if config.MinConn < minMinConn {
		return nil, fmt.Errorf("too little MinConn: %d < %d", config.MinConn, minMinConn)
	}
	if config.MaxConn < config.MinConn {
		return nil, fmt.Errorf("too little MaxConn: %d < MinConn: %d", config.MaxConn, config.MinConn)
	}
	pool = &ConnPool{}
	pool.config = config
	for i := 0;i < config.MinConn;i++ {
		conn, err := pool.makeConn()
		if err != nil {
			return nil, err
		}
		conn.Close()
	}
	return pool, nil
}

func (this *ConnPool) makeConn() (net.Conn, error) {
	if (this.count >= this.config.MaxConn) {
		return nil, fmt.Errorf("reach maxConns: %d", this.count)
	}
	conn, err := net.DialTimeout("tcp", this.config.Addr, this.config.ConnectTimeout)
	if err != nil {
		return nil, err
	}
	this.count++
	return newPConn(conn, this), nil
}

func (this *ConnPool) Get() (conn net.Conn,err error) {
	this.lock.Lock()
	defer this.lock.Unlock()
	connI := this.conns.Front()
	if (connI == nil) {
		return this.makeConn()
	}
	this.conns.Remove(connI)
	return connI.Value.(net.Conn), nil
}

func (this *ConnPool) put(conn *pConn) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.conns.PushBack(conn)
}
