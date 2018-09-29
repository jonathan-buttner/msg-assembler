package main

import "testing"

func TestNewMsgHandler(t *testing.T) {
	handler := NewMsgHandler()
	if handler == nil {
		t.Error("NewMsgHandler returned nil")
	}
}

func TestAddMsgFragment(t *testing.T) {

}
