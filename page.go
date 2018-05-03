package main

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/juju/errors"
)

type Page struct {
	// Start with file header
	fheader FileHeader
	// File space header parts (xdes)
	fspheader FSPHeader
	XDeses    []*XdesEntry
	// Index page part
	// page directory slots
	pheader        PageIndexHeader
	directorySlots []byte
	dslots         []*DSlots
	// Inode page part
	inode INode
	// checksum && lsn
	trailer [8]byte
	// not innodb data
	no     int
	offset int
	pksize int
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

func (p *Page) parse(data []byte, options *parsePageOptions) error {
	var err error
	r := bytes.NewReader(data)
	// Parse file header
	if err = p.fheader.parse(r); nil != err {
		return err
	}

	if p.fheader.typ == pageTypeIndex {
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
	} else if p.fheader.typ == pageTypeFspHDR {
		if err = p.fspheader.parse(r); nil != err {
			return errors.Trace(err)
		}
		if err = p.parseXdeses(r); nil != err {
			return errors.Trace(err)
		}
	} else if p.fheader.typ == pageTypeINode {
		if err = p.inode.parse(r); nil != err {
			return errors.Trace(err)
		}
	}

	// Parse file trailer, last 8 bytes
	copy(p.trailer[:], data[len(data)-8:])

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
