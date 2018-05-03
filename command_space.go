package main

import (
	"bytes"
	"fmt"
	"os"
	"spf13/cobra"
)

type spaceOptions struct {
	file      string
	extend    int
	pageState bool
	unused    bool
	list      bool
}

func newSpaceCommand() *cobra.Command {
	var options spaceOptions
	c := &cobra.Command{
		Use:   "space",
		Short: "show table space segment and extend information",
		Long:  "show table space segment and extend information",
		Run: func(cmd *cobra.Command, args []string) {
			doSpace(cmd, &options)
		},
	}

	c.Flags().StringVarP(&options.file, "file", "f", "", "innodb table space file path")
	c.Flags().IntVarP(&options.extend, "extend", "e", -1, "extend id")
	c.Flags().BoolVarP(&options.pageState, "pstate", "p", false, "show page state")
	c.Flags().BoolVarP(&options.unused, "unused", "u", false, "show unused extend")
	c.Flags().BoolVarP(&options.list, "list", "l", false, "show extend list")

	return c
}

func doSpace(cmd *cobra.Command, options *spaceOptions) {
	if "" == options.file {
		fmt.Println("No input file specified")
		return
	}

	f, err := os.Open(options.file)
	if nil != err {
		fmt.Println("Open file error ", err)
		return
	}
	defer f.Close()

	pages, err := parseInnodbDataFile(f, &parsePageOptions{})
	if nil != err {
		fmt.Println("Parse innodb data file error ", err)
	}
	// Show the page information
	pageCnt := 0
	for pi, page := range pages {
		if page.fheader.typ != pageTypeFspHDR &&
			page.fheader.typ != pageTypeXdes {
			continue
		}

		fmt.Printf("\t\t\t==========PAGE %d OFFSET 0x%04X==========\r\n",
			pi, page.offset)
		// FSP
		if page.fspheader.spaceID != 0 {
			// Header
			fmt.Printf("%-10s", "space id")
			fmt.Printf("%-11s", "page allo")
			fmt.Printf("%-11s", "page init")
			fmt.Printf("%-8s", "flags")
			fmt.Printf("%-15s", "page used(fg)")
			fmt.Printf("%-51s", "free_frag list")
			fmt.Printf("%-51s", "free list")
			fmt.Printf("%-51s", "full_frag list")
			fmt.Printf("%-17s", "next segment id")
			fmt.Printf("%-51s", "full inodes")
			fmt.Printf("%-51s", "free inodes")
			fmt.Printf("\r\n")
			// Columns
			fmt.Printf("%-10d", page.fspheader.spaceID)
			fmt.Printf("%-11d", page.fspheader.highestPageNumberInFile)
			fmt.Printf("%-11d", page.fspheader.highestPageNumberInitialized)
			fmt.Printf("0x%-6.04X", page.fspheader.Flags)
			fmt.Printf("%-15d", page.fspheader.pagesUsedInFreeFrag)
			fmt.Printf("%-51s", page.fspheader.freeFragList.toString(8))
			fmt.Printf("%-51s", page.fspheader.freeList.toString(8))
			fmt.Printf("%-51s", page.fspheader.fullFragList.toString(8))
			fmt.Printf("%-17d", page.fspheader.nextUnusedSegmentID)
			fmt.Printf("%-51s", page.fspheader.fullInodesList.toString(38))
			fmt.Printf("%-51s", page.fspheader.freeInodesList.toString(38))
			fmt.Printf("\r\n\r\n")
		}
		// Xdes
		fmt.Printf("%-13s", "extend")
		fmt.Printf("%-20s", "page range")
		fmt.Printf("%-20s", "file segment id")
		fmt.Printf("%-16s", "state")
		if options.list {
			fmt.Printf("%-37s", "list")
		}
		if options.pageState {
			fmt.Printf("page state (F)ree or (N)ot free")
		}
		fmt.Printf("\r\n")
		for xi, des := range page.XDeses {
			if options.extend >= 0 {
				if xi != options.extend {
					continue
				}
			}
			if !options.unused {
				if des.fileSegmentID == 0 && xi != 0 {
					continue
				}
			}

			extendID := fmt.Sprintf("%d(0x%04X)", xi, 150+xi*40)
			fmt.Printf("%-13s", extendID)
			pageRange := fmt.Sprintf("%d-%d", pageCnt, pageCnt+63)
			fmt.Printf("%-20s", pageRange)
			fmt.Printf("0x%-18.16X", des.fileSegmentID)
			fmt.Printf("0x%-14.08X", des.state)
			if options.list {
				// List ptr is pointer to the prev/next list, so we should adjust the offset
				// Here is the 8 bytes file segment id
				prevOffset := des.list.prevPageOffset
				nextOffset := des.list.nextPageOffset
				if des.list.prevPageNo != 0xffffffff {
					prevOffset -= 8
				}
				if des.list.nextPageNo != 0xffffffff {
					nextOffset -= 8
				}
				liststr := fmt.Sprintf("0x%08X:0x%04X 0x%08X:0x%04X",
					des.list.prevPageNo, prevOffset,
					des.list.nextPageNo, nextOffset)
				fmt.Printf("%-37s", liststr)
			}
			if options.pageState {
				var stateBuf bytes.Buffer
				free := 0
				for i := 0; i < len(des.pageStateBitmap); i++ {
					// Every bytes represents 4 page state (2bit per page)
					tst := des.pageStateBitmap[i]
					for j := 0; j < 4; j++ {
						var mask byte = 0xC0
						mask = mask >> (2 * uint(j))
						lm := 6 - uint(j)*2
						val := (tst & mask)
						if ((val >> lm) & xdesPageStateFree) != 0 {
							stateBuf.WriteString("F")
							free++
						} else {
							stateBuf.WriteString("N")
						}
					}
				}
				stateBuf.WriteString(fmt.Sprintf("(%d free, %d used)", free, 64-free))
				fmt.Printf(stateBuf.String())
			}
			fmt.Printf("\r\n")
			pageCnt += 64
		}
	}
}
