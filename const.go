package main

import "fmt"

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

// Recorder type
const (
	recorderTypeBTreeNode = 0x01
	recorderTypeInfimum   = 0x02
	recorderTypeSupremum  = 0x03
)

// Recorder format
const (
	recorderFormatCompact = iota
	recorderFormatRedundant
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

const (
	xdesPageStateFree = 0x01
)

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
