package main

import (
	"testing"
)

// TestAddMsgFragment tests that the clean up threads remove two messages
func TestAddMsgFragment(t *testing.T) {
	cleanedUp := 0
	fin := make(chan int)
	h := NewMsgHandler(1, func(transID, offset uint32) {
		cleanedUp += 1
		if cleanedUp == 2 {
			fin <- cleanedUp
		}
	})
	f := createValidFrag(false, 0, 0, make([]byte, 100))
	h.AddFragment(f)
	h.AddFragment(createValidFrag(false, 1, 0, make([]byte, 100)))
	// wait for the clean up threads to fire
	<-fin
	if cleanedUp != 2 {
		t.Error("should have cleaned up 2 messages")
	}
}

func TestCompleteMsg(t *testing.T) {
	cleaned := 0
	h := NewMsgHandler(5000, func(t, off uint32) {
		cleaned += 1
	})

	f := createValidFrag(false, 1, 0, make([]byte, 10))
	f2 := createValidFrag(true, 1, 10, make([]byte, 100))
	h.AddFragment(f)
	h.AddFragment(f2)
	if cleaned != 0 {
		t.Error("Shouldn't have cleaned up anything")
	}

}
