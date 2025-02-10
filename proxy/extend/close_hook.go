package extend

import (
	"net"
	"sync"
)

type CloseHookConn struct {
	net.Conn
	closeFunc func()
}

func NewCloseHookConn(conn net.Conn, closeFunc func()) *CloseHookConn {
	return &CloseHookConn{Conn: conn, closeFunc: sync.OnceFunc(closeFunc)}
}

func (c *CloseHookConn) Close() error {
	c.closeFunc()
	return c.Conn.Close()
}

func (c *CloseHookConn) Upstream() any {
	return c.Conn
}

func (c *CloseHookConn) ReaderReplaceable() bool {
	return true
}

func (c *CloseHookConn) WriterReplaceable() bool {
	return true
}
