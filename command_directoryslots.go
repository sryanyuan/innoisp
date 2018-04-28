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
	pksize    int
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
	c.Flags().IntVarP(&options.pksize, "pksize", "k", 8, "primary key size (BIGINT=8,INT=4,SINT=2,TINT=1)")

	return c
}

func doDslots(cmd *cobra.Command, options *dslotsOptions) {
	if "" == options.file {
		fmt.Println("No input file specified")
		return
	}

	if options.pksize != 8 &&
		options.pksize != 4 &&
		options.pksize != 2 &&
		options.pksize != 1 {
		fmt.Println("invalid pksize")
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
		page.pksize = options.pksize
		fmt.Printf("\t\t\t==========PAGE %d OFFSET 0x%04X LEVEL %d==========\r\n",
			pi, page.offset, page.pheader.level)
		fmt.Printf("%-8s%-12s%-12s%-8s%-8s\r\n",
			"slot", "offset", "type", "owned", "key")

		for _, slot := range page.dslots {
			fmt.Printf("%-8d0x%-10.04X%-12s%-8d",
				slot.index, slot.value, slot.typ, slot.owned)
			if nil != slot.rceptr &&
				slot.rceptr.hasKey {
				fmt.Printf("%-8d", slot.rceptr.key)
			} else {
				fmt.Printf("%-8s", "N/A")
			}
			fmt.Printf("\r\n")
			// Show slot reference recorders
			if options.recorders {
				if nil == slot.rcbptr {
					fmt.Printf("No records found\r\n")
				} else {
					fmt.Printf("slot reference: ")
					ptr := slot.rcbptr
					for i := 0; i < int(slot.owned); i++ {
						if i+1 < int(slot.owned) {
							if ptr.hasKey {
								fmt.Printf("[0x%04X K%d]->", ptr.fieldDataOffset, ptr.key)
							} else {
								fmt.Printf("[0x%04X]->", ptr.fieldDataOffset)
							}
						} else {
							if slot.rctype == recorderTypeInfimum {
								fmt.Printf("[infimum own ")
							} else if slot.rctype == recorderTypeSupremum {
								fmt.Printf("[supremum own ")
							} else {
								fmt.Printf("[normal own ")
							}
							if ptr.hasKey {
								fmt.Printf("%d 0x%04X K%d]", slot.owned, ptr.fieldDataOffset, ptr.key)
							} else {
								fmt.Printf("%d 0x%04X]", slot.owned, ptr.fieldDataOffset)
							}
						}
						ptr = ptr.next
					}
					fmt.Printf("\r\n")
				}
			}
		}
	}
}
