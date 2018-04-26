package main

import (
	"io"
	"os"
)

type parsePageOptions struct {
	parseRecords bool
}

func parseInnodbDataFile(f *os.File, options *parsePageOptions) ([]*Page, error) {
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
		if err = page.parse(pageData[:], options); nil != err {
			return nil, err
		}
		pages = append(pages, &page)
		offset += n
	}

	return pages, nil
}
