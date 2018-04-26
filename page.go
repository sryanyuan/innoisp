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
		return fmt.Sprintf("UNKNOWN(0x%04X)", typ)
	}
	return pageTypeStrs[idx]
}

type Page struct {
	// Start with file header
	fheader FileHeader
	pheader PageHeader
	// page directory slots
	directorySlots []byte
	dslots         []*DSlots
	// recorders
	recorders []*compactRecorder
	// checksum && lsn
	trailer [8]byte
	// not innodb data
	offset int
}

func (p *Page) printDirectorySlots() {
	if nil == p.directorySlots {
		return
	}
	fmt.Printf("\t\tPage directory slots (%d total):\r\n[", len(p.directorySlots)/2)
	for i := 0; i < len(p.directorySlots); i += 2 {
		v := binary.BigEndian.Uint16(p.directorySlots[i:])
		fmt.Printf("0x%04X", v)
		if i+2 < len(p.directorySlots) {
			fmt.Printf(" ")
		}
	}
	fmt.Printf("]")
	fmt.Printf("\r\n")
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

func (h *PageHeader) printVerbose() {
	fmt.Printf("\t\tPage header:\r\n")
	fmt.Printf("Heap top <0x%04X> ", h.heapTop)
	fmt.Printf("N heap <0x%04X> ", h.nHeap)
	fmt.Printf("Free <0x%04X> ", h.free)
	fmt.Printf("Garbage <0x%04X> ", h.garbage)
	fmt.Printf("Last insert <0x%04X> ", h.lastInsert)
	fmt.Printf("Direction <0x%04X> ", h.direction)
	fmt.Printf("N direction <0x%04X> ", h.nDirection)
	fmt.Printf("N recs <0x%04X> ", h.nRecs)
	fmt.Printf("Index id <0x%016X> ", h.indexID)
	fmt.Printf("\r\n")
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

type DSlots struct {
	index int
	value uint16
	owned uint8
	typ   string
	rcptr *compactRecorder
}

// 40bit = 5bytes
type compactRecorderHeader struct {
	deleteFlag   bool
	minRecFlag   bool
	Owned        uint8
	heapNo       uint16
	recordType   uint8
	nextRecorder uint16
}

func (h *compactRecorderHeader) parse(data []byte) error {
	if len(data) < 5 {
		return io.EOF
	}

	h.deleteFlag = (data[0] & 0x20) != 0
	h.minRecFlag = (data[0] & 0x10) != 0
	h.Owned = data[0] & 0x0f
	h.heapNo |= uint16(data[1]) << 5
	h.heapNo |= uint16(data[2]&0xf8) >> 3
	h.recordType = data[2] & 0x07
	h.nextRecorder = binary.BigEndian.Uint16(data[3:])

	return nil
}

// Variable length table and null masks not parsed ...
type compactRecorder struct {
	// Variable length table
	// Null masks
	// header
	header compactRecorderHeader
	// Field datas ...
	fieldDataOffset uint16

	next *compactRecorder
}

func (p *Page) parse(data []byte, options *parsePageOptions) error {
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

	if err = p.parseDirectorySlot(data); nil != err {
		return err
	}

	if options.parseRecords {
		if err = p.parseRecorders(data); nil != err {
			return err
		}
	}

	// Parse file trailer, last 8 bytes
	copy(p.trailer[:], data[len(data)-8:])

	return nil
}

func (p *Page) parseDirectorySlot(data []byte) error {
	// Parse directory slots
	if p.pheader.nDirSlots != 0 &&
		p.fheader.typ == pageTypeIndex {
		// Every slot occupy 2 bytes
		p.directorySlots = make([]byte, p.pheader.nDirSlots*2)
		p.dslots = make([]*DSlots, 0, p.pheader.nDirSlots)
		copy(p.directorySlots, data[len(data)-8-len(p.directorySlots):])
	}
	if nil == p.directorySlots {
		return nil
	}

	index := 0
	for i := len(p.directorySlots) - 2; i >= 0; i -= 2 {
		var ds DSlots
		ds.index = index
		ds.value = binary.BigEndian.Uint16(p.directorySlots[i:])
		// Check the record
		var crh compactRecorderHeader
		// Record data offset by slot value is the row data, we need the previous head data to get the
		// header
		recordData := data[ds.value-5:]
		if err := crh.parse(recordData); nil != err {
			return err
		}
		ds.owned = crh.Owned
		// Get record type
		ds.typ = "normal"
		if 0x01 == crh.recordType {
			ds.typ = "b+ tree node ptr"
		} else if 0x02 == crh.recordType {
			ds.typ = "infimum"
		} else if 0x03 == crh.recordType {
			ds.typ = "supremum"
		}
		p.dslots = append(p.dslots, &ds)
		index++
	}

	return nil
}

func (p *Page) parseRecorders(data []byte) error {
	if nil == p.dslots || len(p.dslots) == 0 {
		return nil
	}
	total := 0
	for _, v := range p.dslots {
		total += int(v.owned) + 1
	}
	p.recorders = make([]*compactRecorder, 0, total)
	// Get all records with page directory slot
	for _, slot := range p.dslots {
		recordHeadOffset := int(slot.value) - 5
		var prev *compactRecorder

		for i := 0; i < int(slot.owned); i++ {
			var rc compactRecorder
			recordData := data[recordHeadOffset:]
			if err := rc.header.parse(recordData); nil != err {
				return err
			}

			rc.fieldDataOffset = slot.value
			p.recorders = append(p.recorders, &rc)
			// Find the next recorder
			if i+1 < int(slot.owned) {
				recordHeadOffset += int(rc.header.nextRecorder)
			}
			// Set next pointer
			if nil == prev {
				slot.rcptr = &rc
				prev = &rc
			} else {
				prev.next = &rc
				prev = &rc
			}

			if 0 == rc.header.nextRecorder {
				// Tail
				break
			}
		}
	}

	return nil
}

func (p *Page) printFileTrailer() {
	fmt.Printf("\t\tFile trailer:\r\n")
	// Lower 4 bytes is checksum
	checksum := binary.BigEndian.Uint32(p.trailer[0:4])
	// Higher 4 bytes is lsn
	lsn := binary.BigEndian.Uint32(p.trailer[4:8])
	fmt.Printf("Check sum<0x%08X> LSN<%d> ", checksum, lsn)
	fmt.Printf("\r\n")
}
