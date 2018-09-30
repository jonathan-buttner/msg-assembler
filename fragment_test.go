package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
)

func createHdr(flag bool, dataLen uint16, transID uint32, offset uint32) *bytes.Buffer {
	b := &bytes.Buffer{}
	var flags uint16
	if flag {
		flags = 1
	} else {
		flags = 0
	}

	binary.Write(b, binary.BigEndian, flags)
	binary.Write(b, binary.BigEndian, dataLen)
	binary.Write(b, binary.BigEndian, offset)
	binary.Write(b, binary.BigEndian, transID)

	return b
}

func createFrag(flag bool, transID uint32, offset uint32, data []byte, dataErr bool) io.Reader {
	dataLen := uint16(len(data))
	hdr := createHdr(flag, dataLen, transID, offset)
	if dataErr {
		return hdr
	}
	binary.Write(hdr, binary.BigEndian, data)
	return hdr
}

const (
	errFlags = iota
	errDataLen
	errTransID
	errOffset
)

func createHdrErr(errType int) *bytes.Buffer {
	b := &bytes.Buffer{}
	if errType == errFlags {
		return b
	}
	var flags uint16
	flags = 0
	binary.Write(b, binary.BigEndian, flags)
	if errType == errDataLen {
		return b
	}
	var dataLen uint16 = 1
	binary.Write(b, binary.BigEndian, dataLen)

	if errType == errOffset {
		return b
	}
	var offset uint32 = 10
	binary.Write(b, binary.BigEndian, offset)

	if errType == errTransID {
		return b
	}
	var transID uint32 = 5
	binary.Write(b, binary.BigEndian, transID)
	if errType == errOffset {
		return b
	}

	return b
}

// TestCreateFragHeader tests that the CreateFragHeader function
// correctly constructs a FragmentHdr by reading bytes from an io.Reader.
func TestCreateFragHeader(t *testing.T) {
	data := createHdr(false, 100, 1, 5)
	hdr, _ := CreateFragHeader(data)
	if hdr.IsEnd {
		t.Error("IsEnd should have been false")
	}
	if hdr.DataLen != 100 {
		t.Error("DataLen should have been 100")
	}
	if hdr.TransID != 1 {
		t.Error("Transaction ID should have been 1")
	}
	if hdr.Offset != 5 {
		t.Error("Offset should have been 5")
	}
	data = createHdr(true, 0, 0, 0)
	hdr, _ = CreateFragHeader(data)
	if !hdr.IsEnd {
		t.Error("IsEnd should have been true")
	}
}

// TestHdrErrors tests that the CreateFragHeader returns an error when there
// aren't enough bytes in the reader.
func TestHdrErrors(t *testing.T) {
	reader := createHdrErr(errFlags)
	_, err := CreateFragHeader(reader)
	if err == nil {
		t.Error("CreateFragHeader should have failed for reading flags")
	}
	reader = createHdrErr(errDataLen)
	_, err = CreateFragHeader(reader)
	if err == nil {
		t.Error("CreateFragHeader should have failed for reading data len")
	}
	reader = createHdrErr(errTransID)
	_, err = CreateFragHeader(reader)
	if err == nil {
		t.Error("CreateFragHeader should have failed for reading transaction ID")
	}
	reader = createHdrErr(errOffset)
	_, err = CreateFragHeader(reader)
	if err == nil {
		t.Error("CreateFragHeader should have failed for reading offset")
	}
}

// TestCreateFragment tests that the CreateFragment function builds a Fragment
// structure correctly. It also tests that an error is returned if there aren't
// enough bytes to be read from the reader.
func TestCreateFragment(t *testing.T) {
	data := make([]byte, 100)
	reader := createFrag(true, 10, 11, data, false)
	frag, _ := CreateFragment(reader)
	if !bytes.Equal(frag.Data, data) {
		t.Error("Fragment data wasn't correct")
	}
	reader = createFrag(false, 10, 11, data, true)
	frag, err := CreateFragment(reader)
	if err == nil {
		t.Error("expected an error when reading the data field")
	}
	reader = createHdrErr(errFlags)
	frag, err = CreateFragment(reader)
	if err == nil {
		t.Error("expected an error when creating the header")
	}
}
