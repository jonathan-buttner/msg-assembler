package main

import (
	"bufio"
	"net"
	"sync"
	"time"
)

// Server structure handles receiving UDP messages
type Server struct {
	numThreads int
	netPack    NetWrapper
	handler    *MsgHandler
	address    *net.UDPAddr
	quit       chan bool
	wg         *sync.WaitGroup
	readWait   time.Duration
	errChan    chan error
	conn       Conn
}

// Start spins up the requested number of threads and handles the UDP data
func (s *Server) Start() error {
	conn, err := s.netPack.ListenUDP("udp", s.address)
	s.conn = conn
	if err != nil {
		return err
	}
	for i := 0; i < s.numThreads; i++ {
		go s.handleMsgs()
	}
	return nil
}

// Stop shuts down the  server
func (s *Server) Stop() {
	s.quit <- true
	s.wg.Wait()
	close(s.errChan)
	s.conn.Close()
}

// HandleErrors sends any recieved errors from the udp connection to the
// caller to handle.
func (s *Server) HandleErrors(cb func(err error)) {
	for e := range s.errChan {
		cb(e)
	}
}

func (s *Server) handleMsgs() {
	s.wg.Add(1)
	defer s.wg.Done()
	for {
		// This allows the read to break from the blocking call
		// so the thread can check for the quit signal
		s.conn.SetReadDeadline(time.Now().Add(s.readWait))
		select {
		case <-s.quit:
			return
		default:
			bufReader := bufio.NewReader(s.conn)
			// Create the fragment from the udp traffic
			f, err := CreateFragment(bufReader)
			e, ok := err.(net.Error)
			// check for the timeout or other errors
			if err != nil {
				if !ok || !e.Timeout() {
					s.errChan <- err
					continue
				} else { // timeout
					continue
				}
			} else {
				// Handle the fragment
				s.handler.AddFragment(f)
			}
		}
	}
}

// NewServer initializes a Server structure for handling UDP messages
func NewServer(numThreads int,
	network NetWrapper,
	handler *MsgHandler,
	address *net.UDPAddr,
	readWait time.Duration) *Server {
	return &Server{
		numThreads,
		network,
		handler,
		address,
		make(chan bool),
		&sync.WaitGroup{},
		readWait,
		make(chan error, 100),
		nil,
	}
}
