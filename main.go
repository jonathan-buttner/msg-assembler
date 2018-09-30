package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	fmt.Println("Starting Server")
	n := &NetImp{}
	h := NewMsgHandler(1*1000, PrintHoles, nil)
	addr := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 6789}
	s := NewServer(4, n, h, addr, 5*time.Second)
	s.Start()
	// This will loop forever waiting for errors
	s.HandleErrors(func(e error) {
		fmt.Printf("Error: %v\n", e)
	})
}
