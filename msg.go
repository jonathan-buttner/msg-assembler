package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/deadaccurate/msg-assembler/tree"
)

const (
	// Duplicate is returned by AddFragment when the fragment is already
	// contained in the structure
	Duplicate = iota
	// WrongTransID is returned by AddFragment when the fragment being
	// added has the wrong transaction ID as the first fragment that was
	// used to create this Msg
	WrongTransID
	// Success indicates that the AddFragment method successfully added
	// the fragment.
	Success
)

// Msg is the data model for a message received from the client. It
// stores the fragments and is able to reassemble the fragments into
// the full message.
type Msg struct {
	// transID is the unique message ID
	transID uint32
	// fragTree keeps the fragments in order by offset to aid determining
	// if there is a whole when rebuilding occurs. I chose a binary tree
	// so that I easily keep the fragments sorted by the offset. Keeping
	// them sorted allows for easy determining if there are holes.
	fragTree *tree.Tree
	// recvTotal is the current sum of all the received fragments' data portion
	// for a single transation ID. This is used to tell if the entire message
	// has been received.
	recvTotal uint32
	// total is the required total amount of data for a single message. It is
	// the final packet's offset + data length field.
	total uint32
	// receivedEnd is set to true when the end fragment is received. It is really only
	// necessary just in case a fragment with data length 0 is received that is
	// also the end fragment indicating that a message has size 0. That way the
	// Total can have a legitimate size of 0, otherwise I wouldn't be able to
	// determine if the end fragment had been received yet.
	receivedEnd bool
	// fragMap is a map of Offset to a boolean. It allows O(1) access to
	// determine if the received fragment is a duplicate. The boolean portion
	// is unnecessary but I can't just use a static array because I don't know
	// how many fragments I will receive ahead of time.
	fragMap map[uint32]bool
}

// msgCompare is passed to the binary tree to compare two fragments.
func msgCompare(obj1 interface{}, obj2 interface{}) int {
	if obj1.(*Fragment).Offset < obj2.(*Fragment).Offset {
		return -1
	} else if obj1.(*Fragment).Offset > obj2.(*Fragment).Offset {
		return 1
	}
	return 0
}

// NewMsg creates a new message structure and inserts the specified fragment.
func NewMsg(frag *Fragment) *Msg {
	total := uint32(0)
	if frag.IsEnd {
		// TODO check for overflow
		total = frag.Offset + uint32(frag.DataLen)
	}
	tree := tree.NewTree(msgCompare)
	tree.Insert(frag)
	m := &Msg{
		transID:     frag.TransID,
		fragTree:    tree,
		recvTotal:   uint32(frag.DataLen),
		total:       total,
		receivedEnd: frag.IsEnd,
		fragMap:     map[uint32]bool{frag.Offset: true},
	}
	return m
}

// AddFragment attempts to add a fragment to the message. If the fragment is
// a duplicate (a fragment with the same offset was already added), then the
// enum Duplicate is returned. If the fragment has a different transaction ID
// than this message was created with, WrongTransID is returned. Otherwise
// Success is returned.
func (m *Msg) AddFragment(frag *Fragment) int {
	if frag.TransID != m.transID {
		return WrongTransID
	}

	if _, hasIt := m.fragMap[frag.Offset]; hasIt {
		return Duplicate
	}

	if frag.IsEnd {
		m.total = frag.Offset + uint32(frag.DataLen)
		m.receivedEnd = true
	}

	m.recvTotal += uint32(frag.DataLen)
	m.fragMap[frag.Offset] = true
	m.fragTree.Insert(frag)
	return Success
}

// HasAllFrags checks to see if all the fragments have arrived for this message.
// Returns true if all the fragments have arrived and false otherwise.
// NOTE: This assumes that a client will not send malformed packets. A malformed packet
// would be where two fragments overlap for example:
// Fragment 1: offset = 0, data len = 100
// Fragment 2: offset = 10, data len = 5
func (m *Msg) HasAllFrags() bool {
	if !m.receivedEnd || m.recvTotal != m.total {
		return false
	}
	return true
}

// GetHoles uses the Fragment binary tree to determine if there are any
// missing fragments for this message. If a hole is found it calls the
// cb function with the transaction ID for message and
// the offset of the hole and
func (m *Msg) GetHoles(cb func(transID uint32, startHoleOff uint32)) {
	offArr := m.fragTree.InOrderArr()
	for i := 0; i < len(offArr)-1; i++ {
		curFrag := offArr[i].(*Fragment)
		hole := curFrag.Offset + uint32(curFrag.DataLen)
		if hole != offArr[i+1].(*Fragment).Offset {
			cb(m.transID, hole)
		}
	}
}

// GetSha256 calculates the sha256 hash of all the data for the fragments in the
// message.
func (m *Msg) GetSha256() (string, error) {
	if !m.HasAllFrags() {
		return "", errors.New("Message doesn't have all the fragments")
	}
	h := sha256.New()
	fragArr := m.fragTree.InOrderArr()
	for _, f := range fragArr {
		h.Write(f.(*Fragment).Data)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
