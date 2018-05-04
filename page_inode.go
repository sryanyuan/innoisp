package main

import (
	"encoding/binary"
	"io"

	"github.com/juju/errors"
)

const (
	inodesCountInPage     = 85
	inodeEntryMagicNumber = 97937874
)

type INodeEntry struct {
	// The ID of the file segment (FSEG) described by this
	// file segment INODE entry. If the ID is 0, the entry is unused.
	fileSegmentID uint64
	// Exactly like the space’s FREE_FRAG list (in the FSP header),
	// this field stores the number of pages used in the NOT_FULL list as an
	// optimization to be able to quickly calculate the number of free pages
	// in the list without iterating through all extents in the list.
	usedPagesInNotFullList uint32
	// Extents that are completely unused and are allocated to this file segment.
	freeList ListBaseNode
	// Extents with at least one used page allocated to this file
	// segment. When the last free page is used, the extent is moved to the FULL list.
	notFullList ListBaseNode
	// Extents with no free pages allocated to this file segment.
	// If a page becomes free, the extent is moved to the NOT_FULL list.
	fullList ListBaseNode
	// The value 97937874 is stored as a marker that this
	// file segment INODE entry has been properly initialized.
	magicNumber uint32
	// An array of 32 page numbers of pages allocated individually from
	// extents in the space’s FREE_FRAG or FULL_FRAG list of “fragment” extents.
	// Once this array becomes full, only full extents can be allocated to the file segment.
	fragmentArrayEntry [32]uint32
}

func (n *INodeEntry) parse(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &n.fileSegmentID); nil != err {
		return errors.Trace(err)
	}
	if err := binary.Read(r, binary.BigEndian, &n.usedPagesInNotFullList); nil != err {
		return errors.Trace(err)
	}
	if err := n.freeList.parse(r); nil != err {
		return errors.Trace(err)
	}
	if err := n.notFullList.parse(r); nil != err {
		return errors.Trace(err)
	}
	if err := n.fullList.parse(r); nil != err {
		return errors.Trace(err)
	}
	if err := binary.Read(r, binary.BigEndian, &n.magicNumber); nil != err {
		return errors.Trace(err)
	}
	for i := 0; i < 32; i++ {
		if err := binary.Read(r, binary.BigEndian, &n.fragmentArrayEntry[i]); nil != err {
			return errors.Trace(err)
		}
	}
	return nil
}

type INode struct {
	inodePageList ListNode
	inodes        [inodesCountInPage]*INodeEntry
}

func (n *INode) parse(r io.Reader) error {
	if err := n.inodePageList.parse(r); nil != err {
		return errors.Trace(err)
	}
	for i := 0; i < inodesCountInPage; i++ {
		var entry INodeEntry
		if err := entry.parse(r); nil != err {
			return errors.Trace(err)
		}
		n.inodes[i] = &entry
	}
	return nil
}
