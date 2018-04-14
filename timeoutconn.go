package zhttp

import (
	"net"
	"time"
)

func newTimeoutConn(netConn net.Conn, timeout time.Duration) net.Conn {
	return &timeoutConn{netConn, timeout}
}

type timeoutConn struct {
	net.Conn
	timeout time.Duration
}

func (c *timeoutConn) Read(b []byte) (int, error) {
	if c.timeout > 0 {
		err := c.Conn.SetReadDeadline(time.Now().Add(c.timeout))
		if err != nil {
			return 0, err
		}
	}
	return c.Conn.Read(b)
}

func (c *timeoutConn) Write(b []byte) (int, error) {
	if c.timeout > 0 {
		err := c.Conn.SetWriteDeadline(time.Now().Add(c.timeout))
		if err != nil {
			return 0, err
		}
	}
	return c.Conn.Write(b)
}
