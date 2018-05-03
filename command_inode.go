package main

import (
	"fmt"
	"os"
	"spf13/cobra"
)

type inodeOptions struct {
	file          string
	unused        bool
	fragmentArray bool
}

func newInodeCommand() *cobra.Command {
	var options inodeOptions
	c := &cobra.Command{
		Use:   "inode",
		Short: "show file segment inode information",
		Long:  "show file segment inode information",
		Run: func(cmd *cobra.Command, args []string) {
			doInode(cmd, &options)
		},
	}

	c.Flags().StringVarP(&options.file, "file", "f", "", "innodb table space file path")
	c.Flags().BoolVarP(&options.unused, "unused", "u", false, "show unused inode")
	c.Flags().BoolVarP(&options.fragmentArray, "fragment", "r", false, "show fragment array")

	return c
}

func doInode(cmd *cobra.Command, options *inodeOptions) {
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

	pages, err := parseInnodbDataFile(f, &parsePageOptions{
		parsePageTypeFlag: parsePageInode,
	})
	if nil != err {
		fmt.Println("Parse innodb data file error ", err)
	}

	for _, page := range pages {
		fmt.Printf("\t\t\t==========PAGE %d OFFSET 0x%04X==========\r\n",
			page.no, page.offset)

		// Print table headers
		fmt.Printf("%-51s", "page list")
		fmt.Printf("\r\n")
		// Print table columns
		fmt.Printf("%-51s", page.inode.inodePageList.toString(38))
		fmt.Printf("\r\n\r\n")

		// Print table headers
		fmt.Printf("%-20s", "file segment id")
		fmt.Printf("%-10s", "used(nf)")
		fmt.Printf("%-51s", "free list")
		fmt.Printf("%-51s", "not_full list")
		fmt.Printf("%-51s", "full list")
		if options.fragmentArray {
			fmt.Printf("fragment array")
		}
		fmt.Printf("\r\n")

		// Print inode
		for ni, node := range page.inode.inodes {
			if nil == node {
				panic(fmt.Sprintf("nil inode, index = %d", ni))
			}
			// Print table columns
			if 0 == node.fileSegmentID {
				// Unused
				if !options.unused {
					continue
				}
				fmt.Printf("0x%08X:%-9s", 38+12+ni*192, "<unused>")
			} else {
				fmt.Printf("0x%08X:%-9d", 38+12+ni*192, node.fileSegmentID)
			}

			fmt.Printf("%-10d", node.usedPagesInNotFullList)
			fmt.Printf("%-51s", node.freeList.toString(8))
			fmt.Printf("%-51s", node.notFullList.toString(8))
			fmt.Printf("%-51s", node.fullList.toString(8))

			if options.fragmentArray {
				cnt := 0
				for _, v := range node.fragmentArrayEntry {
					if v == 0xffffffff {
						continue
					}
					fmt.Printf("%d ", v)
					cnt++
				}
				if cnt == len(node.fragmentArrayEntry) {
					fmt.Printf("(extend allocate)")
				} else {
					fmt.Printf("(page allocate)")
				}
			}

			fmt.Printf("\r\n")
		}
	}
}
