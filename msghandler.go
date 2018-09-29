package main

import (
	"sync"
	"time"
)

type MsgHandler struct {
	CleanUpTimer *time.Timer
	MsgMap       map[uint32]*Msg
	Lock         *sync.Mutex
}

func NewMsgHandler() *MsgHandler {
	return &MsgHandler{
		CleanUpTimer: nil,
		MsgMap:       make(map[uint32]*Msg),
		Lock:         &sync.Mutex{},
	}
}
