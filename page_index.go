package main

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/juju/errors"
)

type FileSegmentHeader struct {
	inodeSpaceID    uint32
	inodePageNumber uint32
	inodeOffset     uint16
}

func (f *FileSegmentHeader) parse(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &f.inodeSpaceID); nil != err {
		return errors.Trace(err)
	}
	if err := binary.Read(r, binary.BigEndian, &f.inodePageNumber); nil != err {
		return errors.Trace(err)
	}
	if err := binary.Read(r, binary.BigEndian, &f.inodeOffset); nil != err {
		return errors.Trace(err)
	}
	return nil
}

type PageIndexHeader struct {
	nDirSlots    uint16
	heapTop      uint16
	nHeap        uint16
	free         uint16
	garbage      uint16
	lastInsert   uint16
	direction    uint16
	nDirection   uint16
	nRecs        uint16
	maxTrxID     uint64
	level        uint16
	indexID      uint64
	leafInode    FileSegmentHeader
	nonleafInode FileSegmentHeader
}

func (h *PageIndexHeader) printIndex() {
	fmt.Printf("level <%d> ", h.level)
}

func (h *PageIndexHeader) printVerbose() {
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
	fmt.Printf("Leaf inode <0x%08X:0x%04X> ",
		h.leafInode.inodePageNumber, h.leafInode.inodeOffset)
	fmt.Printf("Non-leaf inode <0x%08X:0x%04X> ",
		h.nonleafInode.inodePageNumber, h.nonleafInode.inodeOffset)
	fmt.Printf("\r\n")
}

func (h *PageIndexHeader) parse(r io.Reader) error {
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
	if err = binary.Read(r, binary.BigEndian, &h.nDirection); nil != err {
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
	if err = h.leafInode.parse(r); nil != err {
		return err
	}
	if err = h.nonleafInode.parse(r); nil != err {
		return err
	}

	return nil
}

type DSlots struct {
	index  int
	value  uint16
	owned  uint8
	typ    string
	rctype uint8
	rcbptr *compactRecorder
	rceptr *compactRecorder
}

// Previous bytes is variable length and null flag masks
// 40bit = 5bytes
// Flags (4 bits) + Number of records owned (4 bits) = 1 byte
// Order (13 bits) + record type (3 bits) = 2 bytes
// Next record offset (2 bytes)
type compactRecorderHeader struct {
	// deleted (2) meaning the record is delete-marked
	// (and will be actually deleted by a purge operation in the future).
	deleteFlag bool
	// min_rec (1) meaning this record is the minimum record in a non-leaf level of the B+Tree
	minRecFlag bool
	// The number of records “owned” by the current record in the page directory.
	Owned uint8
	// The order in which this record was inserted into the heap.
	// Heap records (which include infimum and supremum) are numbered from 0.
	// Infimum is always order 0, supremum is always order 1.
	// User records inserted will be numbered from 2.
	heapNo uint16
	// The type of the record, where currently only 4 values are supported:
	// conventional (0), node pointer (1), infimum (2), and supremum (3).
	recordType uint8
	// A relative offset from the current record to the
	// origin of the next record within the page in ascending order by key.
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

	next   *compactRecorder
	offset uint16 // Offset relative to the page
	hasKey bool
	key    int64 // Only support bigint as primary key
	// If is root page, the node should pointer to internal or leaf node
	pageptr uint32
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
		ds.rctype = crh.recordType
		ds.typ = "normal"
		if recorderTypeBTreeNode == crh.recordType {
			ds.typ = "node ptr"
		} else if recorderTypeInfimum == crh.recordType {
			ds.typ = "infimum"
		} else if recorderTypeSupremum == crh.recordType {
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

	var prevRecorder *compactRecorder

	for _, slot := range p.dslots {
		if slot.rctype == recorderTypeInfimum {
			// Infimum recorder, only own it self, process the next slot
			recorderHeadOffset := slot.value - 5
			prevRecorder = &compactRecorder{}
			prevRecorder.pageptr = 0xffffffff
			if err := prevRecorder.header.parse(data[recorderHeadOffset:]); nil != err {
				return err
			}
			prevRecorder.fieldDataOffset = slot.value
			prevRecorder.offset = recorderHeadOffset
			slot.rcbptr = prevRecorder
			slot.rceptr = prevRecorder
		} else {
			// Normal recorders, find the previous slot
			recorderHeadOffset := prevRecorder.header.nextRecorder + prevRecorder.offset

			for i := 0; i < int(slot.owned); i++ {
				rc := &compactRecorder{}
				rc.pageptr = 0xffffffff
				rc.offset = recorderHeadOffset
				if err := rc.header.parse(data[recorderHeadOffset:]); nil != err {
					return err
				}
				rc.fieldDataOffset = recorderHeadOffset + 5
				if rc.header.recordType != recorderTypeInfimum &&
					rc.header.recordType != recorderTypeSupremum {
					// Get value
					if p.pksize == 8 {
						rc.key = int64(binary.BigEndian.Uint64(data[rc.fieldDataOffset:]))
						rc.key &= 0x7fffffffffffffff
					} else if p.pksize == 4 {
						rc.key = int64(binary.BigEndian.Uint32(data[rc.fieldDataOffset:]))
						rc.key &= 0x7fffffff
					} else if p.pksize == 2 {
						rc.key = int64(binary.BigEndian.Uint16(data[rc.fieldDataOffset:]))
						rc.key &= 0x7fff
					} else if p.pksize == 1 {
						rc.key = int64(data[rc.fieldDataOffset])
						rc.key &= 0x7f
					}

					if p.pheader.level != 0 {
						// root or internal page
						rc.pageptr = binary.BigEndian.Uint32(data[int(rc.fieldDataOffset)+p.pksize:])
					}

					rc.hasKey = true
				}
				if nil == slot.rcbptr {
					slot.rcbptr = rc
				}
				prevRecorder.next = rc
				prevRecorder = rc
				recorderHeadOffset = rc.offset + rc.header.nextRecorder
			}

			slot.rceptr = prevRecorder
		}
	}

	return nil
}
