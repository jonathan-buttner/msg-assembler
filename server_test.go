package main

import (
	"errors"
	"io/ioutil"
	"net"
	"testing"
	"time"
)

type FakeConn struct {
	readErr         error
	closeErr        bool
	readDeadlineErr bool
	readBytes       []byte
	deadLineCB      func()
}

func (c *FakeConn) Read(p []byte) (n int, err error) {
	if c.readErr != nil {
		return 0, c.readErr
	}
	copy(p[:], c.readBytes)
	return len(c.readBytes), nil
}

func (c *FakeConn) Close() error {
	if c.closeErr {
		return errors.New("close error")
	}
	return nil
}

func (c *FakeConn) SetReadDeadline(t time.Time) error {
	if c.readDeadlineErr {
		return errors.New("deadline error")
	}
	if c.deadLineCB != nil {
		c.deadLineCB()
	}
	return nil
}

type FakeNet struct {
	listenErr bool
	conn      *FakeConn
}

func createUDPAddr() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 0,
	}
}

func (n *FakeNet) ListenUDP(network string, address *net.UDPAddr) (Conn, error) {
	if n.listenErr {
		return nil, errors.New("listen error")
	}
	return n.conn, nil
}

func TestNewServerErr(t *testing.T) {
	h := NewMsgHandler(5000, nil, nil)
	s := NewServer(4,
		&FakeNet{listenErr: true},
		h,
		createUDPAddr(),
		time.Second)
	if err := s.Start(); err.Error() != "listen error" {
		t.Error("expected listen error")
	}

}

func createFakeConn(data []byte) *FakeConn {
	return &FakeConn{readBytes: data}
}

type FakeError struct {
	isTimeout bool
}

func (e *FakeError) Timeout() bool {
	return e.isTimeout
}

func (e *FakeError) Temporary() bool {
	return false
}

func (e *FakeError) Error() string {
	return ""
}

// TestReadFragment tests that reading a single fragment succeeds
func TestReadFragment(t *testing.T) {
	r := createFrag(false, 1, 0, make([]byte, 100), false)
	data, _ := ioutil.ReadAll(r)
	end := make(chan bool)

	cl := func(transID, off uint32) {
		if off != 100 {
			t.Error("expected offset of 100")
		}
		end <- true
	}
	h := NewMsgHandler(1, cl, nil)
	s := NewServer(1,
		&FakeNet{conn: createFakeConn(data)},
		h,
		createUDPAddr(),
		time.Millisecond)
	s.Start()
	<-end
	s.Stop()
}
