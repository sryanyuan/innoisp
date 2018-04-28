package main

import (
	"encoding/binary"
	"fmt"
	"io"
)

// FileHeader size 38bytes
type FileHeader struct {
	// If mysql < 4.0.14, it is table space id because all table file is in the single file
	// If mysql >= 4.0.14, it is checksum
	spaceOrChecksum uint32
	// Page offset
	offset uint32
	// prev and next has value if the page is index page
	// note the prev and next pointer is point to the SAME LEVEL of the index
	// Previous page pointer
	prev uint32
	// Next page pointer
	next uint32
	// Log sequence number
	lsn uint64
	// Page type
	typ uint16
	// Flush log sequence number
	fileFlushLSN uint64
	// Arch log no or space id
	// For mysql >= 4.0.14 is space id
	archLogNoOrSpaceID uint32
}

func (h *FileHeader) printVerbose() {
	fmt.Printf("\t\tFile header:\r\n")
	fmt.Printf("Type <%d> ", h.typ)
	fmt.Printf("Checksum <0x%08X> ", h.spaceOrChecksum)
	fmt.Printf("Offset <%d> ", h.offset)
	fmt.Printf("Prev <0x%08X> ", h.prev)
	fmt.Printf("Next <0x%08X> ", h.next)
	fmt.Printf("Log sequence number <%d> ", h.lsn)
	fmt.Printf("Space ID <%d> ", h.archLogNoOrSpaceID)
	fmt.Printf("\r\n")
}

func (h *FileHeader) parse(r io.Reader) error {
	var err error

	if err = binary.Read(r, binary.BigEndian, &h.spaceOrChecksum); nil != err {
		return err
	}
	if err = binary.Read(r, binary.BigEndian, &h.offset); nil != err {
		return err
	}
	if err = binary.Read(r, binary.BigEndian, &h.prev); nil != err {
		return err
	}
	if err = binary.Read(r, binary.BigEndian, &h.next); nil != err {
		return err
	}
	if err = binary.Read(r, binary.BigEndian, &h.lsn); nil != err {
		return err
	}
	if err = binary.Read(r, binary.BigEndian, &h.typ); nil != err {
		return err
	}
	if err = binary.Read(r, binary.BigEndian, &h.fileFlushLSN); nil != err {
		return err
	}
	if err = binary.Read(r, binary.BigEndian, &h.archLogNoOrSpaceID); nil != err {
		return err
	}

	return nil
}
