package main

import (
	"io"
	"os"
)

const (
	parsePageAllocated = 1 << iota
	parsePageUndoLog
	parsePageInode
	parsePageIndex
	parsePageFSP
	parsePageXdes
	parsePageBlob
)

const parsePageAll = parsePageFSP |
	parsePageInode |
	parsePageIndex |
	parsePageAllocated |
	parsePageUndoLog |
	parsePageXdes |
	parsePageBlob

type parsePageOptions struct {
	parseRecords      bool
	parsePageTypeFlag uint64
}

func (o *parsePageOptions) canParse(tp int) bool {
	if o.parsePageTypeFlag == parsePageAll {
		return true
	}

	utp := uint64(tp)
	tv := uint64(0)
	switch utp {
	case pageTypeFspHDR:
		{
			tv = parsePageFSP
		}
	case pageTypeINode:
		{
			tv = parsePageInode
		}
	case pageTypeIndex:
		{
			tv = parsePageIndex
		}
	case pageTypeAllocated:
		{
			tv = parsePageAllocated
		}
	case pageTypeXdes:
		{
			tv = parsePageXdes
		}
	case pageTypeBlob:
		{
			tv = parsePageBlob
		}
	case pageTypeUndoLog:
		{
			tv = parsePageUndoLog
		}
	}

	return (tv & o.parsePageTypeFlag) != 0
}

func parseInnodbDataFile(f *os.File, options *parsePageOptions) ([]*Page, error) {
	// Every page is 16k
	var pageData [16 * 1024]byte
	pages := make([]*Page, 0, 128)
	var offset int

	if 0 == options.parsePageTypeFlag {
		options.parsePageTypeFlag = parsePageAll
	}

	pageNo := 0
	for {
		n, err := io.ReadFull(f, pageData[:])
		if nil != err {
			if err == io.EOF {
				// End of file
				break
			}
			return nil, err
		}

		var page Page
		page.offset = offset
		page.no = pageNo

		if err = page.parse(pageData[:], options); nil != err {
			return nil, err
		}

		if options.canParse(int(page.fheader.typ)) {
			pages = append(pages, &page)
		}

		offset += n
		pageNo++
	}

	return pages, nil
}
