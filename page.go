package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	// B+ tree leaf node (store data)
	pageTypeIndex        = 0x45BF
	pageTypeAllocated    = 0x0000
	pageTypeUndoLog      = 0x0002
	pageTypeINode        = 0x0003
	pageTypeIBufFreeList = 0x0004
	pageTypeIBufBitmap   = 0x0005
	pageTypeSys          = 0x0006
	pageTypeTrxSys       = 0x0007
	pageTypeFspHDR       = 0x0008
	pageTypeXdes         = 0x0009
	pageTypeBlob         = 0x000a
)

var pageTypeStrs = []string{
	"Index",
	"Allocated",
	"Undo log",
	"File segment inode",
	"Insert buffer free list",
	"Insert Buffer bit map",
	"Sys",
	"Trx sys",
	"File space header",
	"Xdes",
	"Blob",
}

func pageTypeToString(typ int) string {
	idx := -1

	switch typ {
	case pageTypeIndex:
		{
			idx = 0
		}
	case pageTypeAllocated:
		{
			idx = 1
		}
	case pageTypeUndoLog:
		{
			idx = 2
		}
	case pageTypeINode:
		{
			idx = 3
		}
	case pageTypeIBufFreeList:
		{
			idx = 4
		}
	case pageTypeIBufBitmap:
		{
			idx = 5
		}
	case pageTypeSys:
		{
			idx = 6
		}
	case pageTypeTrxSys:
		{
			idx = 7
		}
	case pageTypeFspHDR:
		{
			idx = 8
		}
	case pageTypeXdes:
		{
			idx = 9
		}
	case pageTypeBlob:
		{
			idx = 10
		}
	}

	if idx < 0 {
		return fmt.Sprintf("UNKNOWN(%04X)", typ)
	}
	return pageTypeStrs[idx]
}

type Page struct {
	// Start with file header
	fheader FileHeader
	pheader PageHeader
	// not innodb data
	offset int
}

type FileHeader struct {
	// If mysql < 4.0.14, it is table space id because all table file is in the single file
	// If mysql >= 4.0.14, it is checksum
	spaceOrChecksum uint32
	// Page offset
	offset uint32
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

type PageHeader struct {
	nDirSlots  uint16
	heapTop    uint16
	nHeap      uint16
	free       uint16
	garbage    uint16
	lastInsert uint16
	direction  uint16
	nDirection uint16
	nRecs      uint16
	maxTrxID   uint64
	level      uint16
	indexID    uint64
	btrSegLeaf [10]byte
	btrSegTop  [10]byte
}

func (h *PageHeader) printIndex() {
	fmt.Printf("level <%d> ", h.level)
}

func (h *PageHeader) parse(r io.Reader) error {
	var err error

	if err = binary.Read(r, binary.BigEndian, &h.nDirSlots); nil != err {
		return err
	}
	if err = binary.Read(r, binary.BigEndian, &h.heapTop); nil != err {
		return err
	}
	if err = binary.Read(r, binary.BigEndian, &h.nHeap); nil != err {
		return err
	}
	if err = binary.Read(r, binary.BigEndian, &h.free); nil != err {
		return err
	}
	if err = binary.Read(r, binary.BigEndian, &h.garbage); nil != err {
		return err
	}
	if err = binary.Read(r, binary.BigEndian, &h.lastInsert); nil != err {
		return err
	}
	if err = binary.Read(r, binary.BigEndian, &h.direction); nil != err {
		return err
	}
	if err = binary.Read(r, binary.BigEndian, &h.nRecs); nil != err {
		return err
	}
	if err = binary.Read(r, binary.BigEndian, &h.maxTrxID); nil != err {
		return err
	}
	if err = binary.Read(r, binary.BigEndian, &h.level); nil != err {
		return err
	}
	if err = binary.Read(r, binary.BigEndian, &h.indexID); nil != err {
		return err
	}
	if _, err = r.Read(h.btrSegLeaf[:]); nil != err {
		return err
	}
	if _, err = r.Read(h.btrSegTop[:]); nil != err {
		return err
	}

	return nil
}

func (p *Page) parse(data []byte) error {
	var err error
	r := bytes.NewReader(data)
	// Parse file header
	if err = p.fheader.parse(r); nil != err {
		return err
	}
	// Page page header
	if err = p.pheader.parse(r); nil != err {
		return err
	}

	return nil
}
