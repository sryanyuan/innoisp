package main

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/juju/errors"
)

// Reference to https://blog.jcole.us/2013/01/04/page-management-in-innodb-space-files/

// ListBaseNode stores the double-linked list length and the first node as prev
// , the last node as next
type ListBaseNode struct {
	length uint32
	ListNode
}

// 49 max
func (n *ListBaseNode) toString(offset uint16) string {
	prevo := n.prevPageOffset
	nexto := n.nextPageOffset
	if n.prevPageNo != 0xffffffff {
		prevo -= offset
	}
	if n.nextPageNo != 0xffffffff {
		nexto -= offset
	}
	return fmt.Sprintf("len<%d> 0x%08X:0x%04X 0x%08X:0x%04X",
		n.length, n.prevPageNo, prevo, n.nextPageNo, nexto)
}

func (n *ListBaseNode) parse(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &n.length); nil != err {
		return errors.Trace(err)
	}
	if err := n.ListNode.parse(r); nil != err {
		return errors.Trace(err)
	}
	return nil
}

// ListNode stores the prev and next pointer
type ListNode struct {
	prevPageNo     uint32
	prevPageOffset uint16
	nextPageNo     uint32
	nextPageOffset uint16
}

func (n *ListNode) parse(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &n.prevPageNo); nil != err {
		return errors.Trace(err)
	}
	if err := binary.Read(r, binary.BigEndian, &n.prevPageOffset); nil != err {
		return errors.Trace(err)
	}
	if err := binary.Read(r, binary.BigEndian, &n.nextPageNo); nil != err {
		return errors.Trace(err)
	}
	if err := binary.Read(r, binary.BigEndian, &n.nextPageOffset); nil != err {
		return errors.Trace(err)
	}
	return nil
}

type FSPHeader struct {
	// The space ID of the current space.
	spaceID uint32
	unused  uint32
	// The “size” is the highest valid page number, and is incremented
	// when the file is grown. However, not all of these pages are initialized
	// (some may be zero-filled), as extending a space is a multi-step process.
	highestPageNumberInFile uint32
	// The “free limit” is the highest page number for which the FIL header has
	// been initialized, storing the page number in the page itself, amongst other things.
	// The free limit will always be less than or equal to the size.
	highestPageNumberInitialized uint32
	// Storage of flags related to the space.
	Flags uint32
	//
	pagesUsedInFreeFrag uint32
	// Extents that are completely unused and available to be allocated in
	// whole to some purpose. A FREE extent could be allocated to a file
	// segment (and placed on the appropriate INODE list), or moved
	// to the FREE_FRAG list for individual page use.
	freeList ListBaseNode
	// Extents with free pages remaining that are allocated to be used in “fragments”,
	// having individual pages allocated to different purposes rather than allocating
	// the entire extent. For example, every extent with an FSP_HDR or XDES page will be
	// placed on the FREE_FRAG list so that the remaining free pages in the extent can be
	// allocated for other uses.
	freeFragList ListBaseNode
	// Exactly like FREE_FRAG but for extents with no free pages remaining. Extents are
	// moved from FREE_FRAG to FULL_FRAG when they become full, and moved back to FREE_FRAG
	// if a page is released so that they are no longer full.
	fullFragList ListBaseNode
	// The file segment ID that will be used for the next allocated file segment.
	// (This is essentially an auto-increment integer.)
	nextUnusedSegmentID uint64
	fullInodesList      ListBaseNode
	freeInodesList      ListBaseNode
}

func (h *FSPHeader) parse(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &h.spaceID); nil != err {
		return errors.Trace(err)
	}
	if err := binary.Read(r, binary.BigEndian, &h.unused); nil != err {
		return errors.Trace(err)
	}
	if err := binary.Read(r, binary.BigEndian, &h.highestPageNumberInFile); nil != err {
		return errors.Trace(err)
	}
	if err := binary.Read(r, binary.BigEndian, &h.highestPageNumberInitialized); nil != err {
		return errors.Trace(err)
	}
	if err := binary.Read(r, binary.BigEndian, &h.Flags); nil != err {
		return errors.Trace(err)
	}
	if err := binary.Read(r, binary.BigEndian, &h.pagesUsedInFreeFrag); nil != err {
		return errors.Trace(err)
	}
	if err := h.freeList.parse(r); nil != err {
		return errors.Trace(err)
	}
	if err := h.freeFragList.parse(r); nil != err {
		return errors.Trace(err)
	}
	if err := h.fullFragList.parse(r); nil != err {
		return errors.Trace(err)
	}
	if err := binary.Read(r, binary.BigEndian, &h.nextUnusedSegmentID); nil != err {
		return errors.Trace(err)
	}
	if err := h.fullInodesList.parse(r); nil != err {
		return errors.Trace(err)
	}
	if err := h.freeInodesList.parse(r); nil != err {
		return errors.Trace(err)
	}

	return nil
}

// XdesEntry describe which pages within the extend are in use
type XdesEntry struct {
	// The ID of the file segment to which the extent belongs,
	// if it belongs to a file segment.
	fileSegmentID uint64
	// Pointers to previous and next extents in a doubly-linked extent descriptor list.
	// 6bytes
	list ListNode
	// State: The current state of the extent, for which only four values are currently
	// defined: FREE, FREE_FRAG, and FULL_FRAG, meaning this extent belongs to the
	// space’s list with the same name; and FSEG, meaning this extent belongs to
	// the file segment with the ID stored in the File Segment ID field. (More on these lists below.)
	// TODO: get the definition of state
	state uint32
	// Page State Bitmap: A bitmap of 2 bits per page in the extent (64 x 2 = 128 bits, or 16 bytes).
	// The first bit indicates whether the page is free. The second bit is reserved to indicate whether
	// the page is clean (has no un-flushed data), but this bit is currently unused and is always set to 1.
	pageStateBitmap [16]byte
}

func (e *XdesEntry) GetPageState(pn int) byte {
	return 0
}

func (e *XdesEntry) parse(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &e.fileSegmentID); nil != err {
		return errors.Trace(err)
	}
	if err := e.list.parse(r); nil != err {
		return errors.Trace(err)
	}
	if err := binary.Read(r, binary.BigEndian, &e.state); nil != err {
		return errors.Trace(err)
	}
	if _, err := r.Read(e.pageStateBitmap[:]); nil != err {
		return errors.Trace(err)
	}
	return nil
}

func (p *Page) parseXdeses(r io.Reader) error {
	p.XDeses = make([]*XdesEntry, 0, 256)
	for i := 0; i < cap(p.XDeses); i++ {
		var des XdesEntry
		if err := des.parse(r); nil != err {
			return errors.Trace(err)
		}
		p.XDeses = append(p.XDeses, &des)
	}
	return nil
}
