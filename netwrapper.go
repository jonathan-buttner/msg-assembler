package main

import (
	"io"
	"net"
	"time"
)

// NetWrapper wraps a couple net calls so I can easily mock them in my
// tests
type NetWrapper interface {
	ListenUDP(network string, address *net.UDPAddr) (Conn, error)
}

// Conn wraps the necessary interface for this server's calls to
// read data and close the connection
// The tests will simply implement this interface to test the network
// code.
type Conn interface {
	io.Reader
	io.Closer
	SetReadDeadline(t time.Time) error
}

// ConnImp actually holds a reference to the UDPConn structure that's
// used under the scenes for the actual implementation.
type ConnImp struct {
	conn *net.UDPConn
}

// Read satisfies the io.Reader interface and simply calls the UDPConn's Read method.
func (c *ConnImp) Read(p []byte) (n int, err error) {
	return c.conn.Read(p)
}

// Close passes the call to the underlying UDPConn's Close method.
func (c *ConnImp) Close() error {
	return c.conn.Close()
}

// SetReadDeadline passes the call to the UDPConn's implementation.
func (c *ConnImp) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// NetImp is the wrapped implementation for the NetWrapper interface
type NetImp struct {
}

// ListenUDP embeds the returned UDPConn in a ConnImp struct
func (n *NetImp) ListenUDP(network string, address *net.UDPAddr) (Conn, error) {
	c, err := net.ListenUDP(network, address)
	if err != nil {
		return nil, err
	}
	return &ConnImp{conn: c}, nil
}
