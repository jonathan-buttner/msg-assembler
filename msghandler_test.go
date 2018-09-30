package main

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// TestAddMsgFragment tests that the clean up threads remove two messages
func TestAddMsgFragment(t *testing.T) {
	cleanedUp := 0
	fin := make(chan int, 1)
	h := NewMsgHandler(1, func(transID, offset uint32) {
		cleanedUp++
		if cleanedUp == 2 {
			fin <- cleanedUp
		}
	}, nil)
	f := createValidFrag(false, 0, 0, make([]byte, 100))
	h.AddFragment(f)
	h.AddFragment(createValidFrag(false, 1, 0, make([]byte, 100)))
	// wait for the clean up threads to fire
	<-fin
	if cleanedUp != 2 {
		t.Error("should have cleaned up 2 messages")
	}
}

// TestCompleteMsg tests that the message handler will correctly identify when all
// the fragments are received for a message
func TestCompleteMsg(t *testing.T) {
	cleaned := 0
	rebuilt := make(chan int, 1)
	numRebuilt := 0

	cleanCB := func(t, off uint32) {
		cleaned++
	}
	rebuildCB := func(transID uint32, sha string) {
		if transID != 1 {
			t.Error("Should have rebuilt transID 1")
		}
		numRebuilt++
		rebuilt <- 1
	}
	h := NewMsgHandler(5000, cleanCB, rebuildCB)

	f := createValidFrag(false, 1, 0, make([]byte, 10))
	f2 := createValidFrag(true, 1, 10, make([]byte, 100))
	h.AddFragment(f)
	h.AddFragment(f2)
	if cleaned != 0 {
		t.Error("Shouldn't have cleaned up anything")
	}
	<-rebuilt
	if numRebuilt != 1 {
		t.Errorf("Should have rebuilt 1 message, rebuilt: %d\n", numRebuilt)
	}
}

// TestCorrectSha tests that the calculated sha256 hash is the expected one
// after reassembling fragments for a message
func TestCorrectSha(t *testing.T) {
	rebuilt := make(chan int, 1)
	shaHash := sha256.New()
	data := make([]byte, 100)
	shaHash.Write(data)
	sh := hex.EncodeToString(shaHash.Sum(nil))
	rebCB := func(transID uint32, sha string) {
		if sha != sh {
			t.Error("hashes didn't match")
		}
		rebuilt <- 1
	}
	h := NewMsgHandler(5000, PrintHoles, rebCB)
	f := createValidFrag(true, 1, 0, data)
	h.AddFragment(f)
	<-rebuilt
}

// TestCleanUpAnomaly tests the anomaly case where the clean up task wasn't
// created when the first fragment of a message is recieved
func TestCleanUpAnomaly(t *testing.T) {
	clFun := func(transID, off uint32) {
	}
	h := NewMsgHandler(5000, clFun, nil)
	f := createValidFrag(false, 1, 0, make([]byte, 100))
	h.AddFragment(f)
	h.lock.Lock()
	c, _ := h.cleanUpMap[1]
	c.cleanUpTimer.Stop()
	delete(h.cleanUpMap, 1)
	h.lock.Unlock()
	f = createValidFrag(false, 1, 100, make([]byte, 10))
	h.AddFragment(f)

	if _, ok := h.cleanUpMap[1]; !ok {
		t.Error("clean up msg entry should have been added")
	}
}
