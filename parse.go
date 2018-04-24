package main

import (
	"fmt"
	"io"
	"os"
)

func parseInnodbDataFile(f *os.File) ([]*Page, error) {
	// Every page is 16k
	var pageData [16 * 1024]byte
	pages := make([]*Page, 0, 128)
	var offset int

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
		if err = page.parse(pageData[:]); nil != err {
			return nil, err
		}
		pages = append(pages, &page)
		offset += n
	}

	return pages, nil
}

func printPages(pages []*Page) {
	for i, page := range pages {
		fmt.Printf("==========PAGE %d==========\r\n", i)
		fmt.Printf("page num %d, offset %08X, ", i, page.offset)
		fmt.Printf("page type <%s> ", pageTypeToString(int(page.fheader.typ)))
		if page.fheader.typ == pageTypeIndex {
			page.pheader.printIndex()
		}
		fmt.Printf("\r\n")
	}
}
