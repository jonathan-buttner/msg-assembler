package main

import (
	"testing"
)

func createValidFrag(isEnd bool, tID uint32, offset uint32, data []byte) *Fragment {
	f, _ := CreateFragment(createFrag(isEnd, tID, offset, data, false))
	return f
}

// TestNewMsg tests to make sure that the NewMsg function creates a Msg structure
// and initializes it correctly
func TestNewMsg(t *testing.T) {
	data := make([]byte, 100)
	f := createValidFrag(false, 5, 1, data)
	m := NewMsg(f)
	if m.transID != 5 {
		t.Error("trans ID was not 5")
	}
	if _, ok := m.fragMap[1]; !ok {
		t.Error("FragMap didn't contain the offset")
	}
	if m.total != 0 {
		t.Error("Total should have been initialized to zero")
	}
	if m.receivedEnd {
		t.Error("Shouldn't have received the end fragment yet")
	}
	if m.recvTotal != 100 {
		t.Error("RecvTotal wasn't data + offset")
	}
	arr := m.fragTree.InOrderArr()
	if arr[0].(*Fragment).Offset != 1 {
		t.Errorf("Fragment wasn't inserted into the tree %v\n", arr[0])
	}
}

// TestNewMsgEnd tests NewMsg when a end fragment is received first for the
// transaction ID.
func TestNewMsgEnd(t *testing.T) {
	data := make([]byte, 100)
	f := createValidFrag(true, 5, 1, data)
	m := NewMsg(f)
	if m.total != 101 {
		t.Error("Total should have been set for end fragment")
	}
	if !m.receivedEnd {
		t.Error("Should have received end fragment")
	}
}

// TestMsgCompare tests the msgCompare function used for comparisons in the
// binary tree.
func TestMsgCompare(t *testing.T) {
	data := make([]byte, 10)
	obj1 := createValidFrag(false, 1, 100, data)
	obj2 := createValidFrag(false, 1, 100, data)
	// test equal
	if v := msgCompare(obj1, obj2); v != 0 {
		t.Error("Should have been equal")
	}
	// less than
	obj2 = createValidFrag(false, 1, 99, data)
	if v := msgCompare(obj1, obj2); v != 1 {
		t.Error("Should have been great than")
	}

	// greater than
	obj2 = createValidFrag(false, 1, 101, data)
	if v := msgCompare(obj1, obj2); v != -1 {
		t.Error("Should have been less than")
	}
}

// TestMsgAddFragment tests the AddFragment method by verifying
// that the binary tree has all the fragments in it.
func TestMsgAddFragment(t *testing.T) {
	data := make([]byte, 5)
	// bytes 100 - 105
	f := createValidFrag(false, 1, 100, data)
	m := NewMsg(f)
	// bytes 0 - 100
	f = createValidFrag(false, 1, 0, make([]byte, 100))
	m.AddFragment(f)
	// bytes 105 - 115
	f = createValidFrag(false, 1, 105, make([]byte, 10))
	if s := m.AddFragment(f); s != Success {
		t.Error("AddFragment should have succeeded")
	}
	// tree should be:
	//       100
	//      /   \
	//     0    105
	arr := m.fragTree.InOrderArr()
	if arr[0].(*Fragment).Offset != 0 ||
		arr[1].(*Fragment).Offset != 100 || arr[2].(*Fragment).Offset != 105 {
		t.Errorf("Tree returned an unsorted array %v\n", arr)
	}
}

func createMsgHelper() *Msg {
	f := createValidFrag(false, 1, 0, make([]byte, 10))
	return NewMsg(f)
}

// TestMsgAddFragmentDup tests that adding a duplicate fragment isn't added.
func TestMsgAddFragmentDup(t *testing.T) {
	m := createMsgHelper()
	if ret := m.AddFragment(createValidFrag(false, 1, 0, make([]byte, 100))); ret != Duplicate {
		t.Error("expected duplicate")
	}
}

// TestMsgAddFragWrongID tests that adding a fragment with the wrong transaction ID fails.
func TestMsgAddFragWrongID(t *testing.T) {
	m := createMsgHelper()
	if ret := m.AddFragment(createValidFrag(false, 2, 0, make([]byte, 100))); ret != WrongTransID {
		t.Error("expected wrong transaction ID")
	}
}

func createCompleteMsg() *Msg {
	m := createMsgHelper()
	f := createValidFrag(true, 1, 10, make([]byte, 100))
	m.AddFragment(f)
	return m
}

// TestMsgAddFragEnd tests that adding the end fragment sets the correct fields
func TestMsgAddFragEnd(t *testing.T) {
	m := createCompleteMsg()
	if !m.receivedEnd {
		t.Error("should have received the end fragment")
	}
	if m.total != 110 {
		t.Error("expected total to be 110")
	}
	if m.recvTotal != 110 {
		t.Error("expected received amount to be 110")
	}
}

// TestMsgHasAllFrags tests the HasAllFrags method for whether it returns true when
// all frags were received and false if not.
func TestMsgHasAllFrags(t *testing.T) {
	m := createCompleteMsg()
	if !m.HasAllFrags() {
		t.Error("expected message to have all fragments")
	}
	m = createMsgHelper()
	if m.HasAllFrags() {
		t.Error("message should have received all fragments yet")
	}
}

// TestMsgGetHoles tests that the GetHoles method correctly calls the callback
// parameter with any holes within the message. A hole is an unreceived fragment.
func TestMsgGetHoles(t *testing.T) {
	m := createMsgHelper()
	m.AddFragment(createValidFrag(false, 1, 50, make([]byte, 50)))
	m.AddFragment(createValidFrag(true, 1, 200, make([]byte, 100)))
	holes := make([]uint32, 0)
	m.GetHoles(func(transID uint32, startHoleOff uint32) {
		holes = append(holes, startHoleOff)
		if transID != 1 {
			t.Error("expected trans ID to be 1")
		}
	})
	if len(holes) != 2 {
		t.Error("expected two holes")
	}
	if holes[0] != 10 {
		t.Error("expected hole at offset 50")
	}
	if holes[1] != 100 {
		t.Error("expected hole at offset 50")
	}
}

func createCompleteMsgUnOrdered() *Msg {
	// 5 - 6
	fEnd := createValidFrag(true, 1, 5, make([]byte, 1))
	// 2 - 3
	f2 := createValidFrag(false, 1, 2, make([]byte, 1))
	// 3 - 5
	f3 := createValidFrag(false, 1, 3, make([]byte, 2))
	// 0 - 2
	f1 := createValidFrag(false, 1, 0, make([]byte, 2))
	m := NewMsg(fEnd)
	m.AddFragment(f2)
	m.AddFragment(f1)
	m.AddFragment(f3)
	return m
}

// TestMsgNoHoles builds a message in an unordered way and tests
// to make sure that the message had no holes.
func TestMsgNoHoles(t *testing.T) {
	m := createCompleteMsgUnOrdered()
	m.GetHoles(func(transID uint32, startHoleOff uint32) {
		t.Error("There should be no holes")
	})
}

func TestMsgGetSha256(t *testing.T) {
	m := createCompleteMsgUnOrdered()
	sh, _ := m.GetSha256()
	expSha := "b0f66adc83641586656866813fd9dd0b8ebb63796075661ba45d1aa8089e1d44"
	if sh != expSha {
		t.Errorf("calculated sha256: %s\nexp: %s\n", sh, expSha)
	}
	m = createMsgHelper()
	sh, err := m.GetSha256()
	if sh != "" || err == nil {
		t.Error("expected an error since all the fragments haven't arrived")
	}
}
