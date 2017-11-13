package huskarPool

import (
	"net"
	"sync"
)

type CloseHook struct {
	net.Conn
	once       sync.Once
	afterClose func()
}

func HookClose(conn net.Conn, closeCallback func()) net.Conn {
	return &CloseHook{
		Conn:       conn,
		afterClose: closeCallback,
	}
}

func (c *CloseHook) Close() error {
	err := c.Conn.Close()
	if c.afterClose != nil {
		c.once.Do(c.afterClose)
	}
	return err
}
