package main

import (
	"encoding/binary"
	"fmt"
	"io"
)

type PageIndexHeader struct {
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
	index  int
	value  uint16
	owned  uint8
	typ    string
	rctype uint8
	rcbptr *compactRecorder
	rceptr *compactRecorder
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

	next   *compactRecorder
	offset uint16 // Offset relative to the page
	hasKey bool
	key    int64 // Only support bigint as primary key
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
