package main

import (
	"fmt"
	"os"
	"spf13/cobra"
)

type overviewOptions struct {
	file    string
	verbose bool
	page    int
}

func newOverviewCommand() *cobra.Command {
	var options overviewOptions
	c := &cobra.Command{
		Use:   "overview",
		Short: "overview table space file",
		Long:  "Overview the innodb table space file(*.ibd)",
		Run: func(cmd *cobra.Command, args []string) {
			doOverview(cmd, &options)
		},
	}

	c.Flags().StringVarP(&options.file, "file", "f", "", "innodb table space file path")
	c.Flags().BoolVarP(&options.verbose, "verbose", "v", false, "show verbose information")
	c.Flags().IntVarP(&options.page, "page", "p", -1, "specify page to show")

	return c
}

func doOverview(cmd *cobra.Command, options *overviewOptions) {
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
	printInnodbPages(pages, options)
}

func printInnodbPages(pages []*Page, options *overviewOptions) {
	for i, page := range pages {
		if options.page >= 0 {
			if i != options.page {
				continue
			}
		}
		fmt.Printf("==========PAGE %d==========\r\n", i)
		fmt.Printf("page num %d, offset 0x%08X, ", i, page.offset)
		fmt.Printf("page type <%s> ", pageTypeToString(int(page.fheader.typ)))
		if page.fheader.typ == pageTypeIndex {
			page.pheader.printIndex()
		}
		fmt.Printf("\r\n")
		if options.verbose {
			page.fheader.printVerbose()
			page.pheader.printVerbose()
			page.printFileTrailer()
			page.printDirectorySlots()
		}
		fmt.Printf("\r\n")
	}
}
