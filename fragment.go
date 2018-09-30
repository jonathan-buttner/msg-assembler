package main

import (
	"encoding/binary"
	"io"
)

// FragmentHdr defines the header portion of a fragmented packet
type FragmentHdr struct {
	IsEnd   bool
	DataLen uint16
	TransID uint32
	Offset  uint32
}

// CreateFragHeader reads from the reader and creates a fragment header.
func CreateFragHeader(reader io.Reader) (*FragmentHdr, error) {
	hdr := &FragmentHdr{}
	var flags uint16
	var err error
	if err = binary.Read(reader, binary.BigEndian, &flags); err != nil {
		return nil, err
	}

	if flags > 0 {
		hdr.IsEnd = true
	} else {
		hdr.IsEnd = false
	}

	if err = binary.Read(reader, binary.BigEndian, &hdr.DataLen); err != nil {
		return nil, err
	}
	if err = binary.Read(reader, binary.BigEndian, &hdr.Offset); err != nil {
		return nil, err
	}
	if err = binary.Read(reader, binary.BigEndian, &hdr.TransID); err != nil {
		return nil, err
	}

	return hdr, nil
}

// Fragment represents a fragment header and a portion of a full message which
// is stored in the data field
type Fragment struct {
	FragmentHdr
	Data []byte
}

// CreateFragment reads from the reader and creates a full Fragment object.
// It returns an error if there wasn't enough bytes to create the fragment.
func CreateFragment(reader io.Reader) (*Fragment, error) {
	frag := &Fragment{}

	hdr, err := CreateFragHeader(reader)
	if err != nil {
		return nil, err
	}
	frag.FragmentHdr = *hdr
	data := make([]byte, frag.DataLen)
	if err = binary.Read(reader, binary.BigEndian, data); err != nil {
		return nil, err
	}
	frag.Data = data
	return frag, nil
}
