package main

import (
	"fmt"
	"sync"
	"time"
)

type cleanUpMsg struct {
	cleanUpTimer *time.Timer
	msgHandler   *MsgHandler
	transID      uint32
}

func (c *cleanUpMsg) cleanUp() {
	c.msgHandler.lock.Lock()
	defer c.msgHandler.lock.Unlock()
	// if the transaction ID is no longer in the map then
	// the message was reassembled and removed while we were waiting for the lock
	if m, ok := c.msgHandler.msgMap[c.transID]; ok {
		// only clean up if we don't have all the fragments
		// if a fragment sunk in just in time let the reassembly happen
		if !m.HasAllFrags() {
			delete(c.msgHandler.msgMap, c.transID)
			delete(c.msgHandler.cleanUpMap, c.transID)
			// call the callback so the holes can be printed
			m.GetHoles(c.msgHandler.cleanUpCB)
		}
	}
}

// MsgHandler handles locking and cleanup for messages. It allows fragments to be
// added to messages.
type MsgHandler struct {
	cleanUpDelay int
	cleanUpCB    func(transID, offset uint32)
	cleanUpMap   map[uint32]*cleanUpMsg
	msgMap       map[uint32]*Msg
	lock         *sync.Mutex
	rebuiltMsgCB func(transID uint32, sha256 string)
}

// NewMsgHandler creates a MsgHandler. The MsgHandler handles thread safety for
// making storing fragments. It also deletes a message after the specified time.
// cleanUpWait is the number of milliseconds to wait before removing an entry. cb
// is called after the message is removed.
func NewMsgHandler(cleanUpWait int,
	cleanUpCB func(uint32, uint32),
	rebuiltCB func(uint32, string)) *MsgHandler {
	h := &MsgHandler{
		cleanUpCB:    cleanUpCB,
		cleanUpDelay: cleanUpWait,
		cleanUpMap:   make(map[uint32]*cleanUpMsg),
		msgMap:       make(map[uint32]*Msg),
		lock:         &sync.Mutex{},
		rebuiltMsgCB: rebuiltCB,
	}
	return h
}

// PrintHoles is a callback for when the cleanup thread removes the fragments
// for a message. This function provides a default implementation for the callback
// which prints the holes.
func PrintHoles(transID, off uint32) {
	fmt.Printf("Message #%d Hole at: %d\n", transID, off)
}

func (h *MsgHandler) addCleanUpMsg(transID uint32) *cleanUpMsg {
	clMsg := &cleanUpMsg{
		cleanUpTimer: nil,
		msgHandler:   h,
		transID:      transID,
	}
	dur := time.Duration(h.cleanUpDelay) * time.Millisecond
	// start the clean up timer
	clMsg.cleanUpTimer = time.AfterFunc(dur, clMsg.cleanUp)
	h.cleanUpMap[transID] = clMsg
	return clMsg
}

func (h *MsgHandler) reassembleMsg(msg *Msg) {
	sh, _ := msg.GetSha256()
	if h.rebuiltMsgCB != nil {
		h.rebuiltMsgCB(msg.transID, sh)
	}
	fmt.Printf("Message #%d length: %d\n", msg.transID, msg.total)
	fmt.Printf("sha256:%s\n", sh)
}

// AddFragment handles thread safety and clean up of an incomplete message when
// it hasn't arrived after the specified wait time. To add a fragment pass it
// to this method.
func (h *MsgHandler) AddFragment(frag *Fragment) {
	h.lock.Lock()
	defer h.lock.Unlock()
	var msg *Msg
	var clMsg *cleanUpMsg
	// message trans ID exists in the map
	if msgInMap, ok := h.msgMap[frag.TransID]; ok {
		msgInMap.AddFragment(frag)
		clMsg, ok = h.cleanUpMap[frag.TransID]
		// this is an anomaly! It should have already been the map
		if !ok {
			clMsg = h.addCleanUpMsg(frag.TransID)
		}
		msg = msgInMap
	} else { // message trans id didn't exist so add it and set clean up timer
		msg = NewMsg(frag)
		h.msgMap[frag.TransID] = msg
		clMsg = h.addCleanUpMsg(frag.TransID)
	}

	if msg.HasAllFrags() {
		delete(h.msgMap, frag.TransID)
		delete(h.cleanUpMap, frag.TransID)
		clMsg.cleanUpTimer.Stop()
		h.reassembleMsg(msg)
	}
}
