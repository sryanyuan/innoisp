package main

import (
	"fmt"
	"os"
	"spf13/cobra"
)

type dslotsOptions struct {
	file      string
	page      int
	recorders bool
}

func newDslotsCommand() *cobra.Command {
	var options dslotsOptions
	c := &cobra.Command{
		Use:   "dslots",
		Short: "directory slots",
		Long:  "Show the directory slots information",
		Run: func(cmd *cobra.Command, args []string) {
			doDslots(cmd, &options)
		},
	}

	c.Flags().StringVarP(&options.file, "file", "f", "", "innodb table space file path")
	c.Flags().IntVarP(&options.page, "page", "p", -1, "specify page to show directory slots")
	c.Flags().BoolVarP(&options.recorders, "recorders", "r", false, "show slot reference recorders")

	return c
}

func doDslots(cmd *cobra.Command, options *dslotsOptions) {
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
		parseRecords: options.recorders,
	})
	if nil != err {
		fmt.Println("Parse innodb data file error ", err)
	}
	// Show the page information
	for pi, page := range pages {
		if options.page >= 0 {
			if options.page != pi {
				continue
			}
		}
		// Only index page have dslots
		if page.dslots == nil {
			continue
		}
		fmt.Printf("\t\t\t==========PAGE %d==========\r\n", pi)
		fmt.Printf("%-8s%-12s%-12s%-8s%-8s\r\n",
			"slot", "offset", "type", "owned", "key")

		for _, slot := range page.dslots {
			fmt.Printf("%-8d0x%-10.04X%-12s%-8d\r\n",
				slot.index, slot.value, slot.typ, slot.owned)
			// Show slot reference recorders
			if options.recorders {
				if nil == slot.rcptr {
					fmt.Printf("No records found\r\n")
				} else {
					fmt.Printf("slot reference -> ")
					ptr := slot.rcptr
					for nil != ptr {
						fmt.Printf("[0x%04X]", ptr.fieldDataOffset)
						ptr = ptr.next
						if nil != ptr {
							fmt.Printf(" ->")
						}
					}
					fmt.Printf("\r\n")
				}
			}
		}
	}
}
