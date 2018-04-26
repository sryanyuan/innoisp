package main

import (
	"fmt"
	"os"
	"spf13/cobra"
)

type overviewOptions struct {
	file    string
	verbose bool
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
	printInnodbPages(pages, options.verbose)
}

func printInnodbPages(pages []*Page, verbose bool) {
	for i, page := range pages {
		fmt.Printf("==========PAGE %d==========\r\n", i)
		fmt.Printf("page num %d, offset 0x%08X, ", i, page.offset)
		fmt.Printf("page type <%s> ", pageTypeToString(int(page.fheader.typ)))
		if page.fheader.typ == pageTypeIndex {
			page.pheader.printIndex()
		}
		fmt.Printf("\r\n")
		if verbose {
			page.fheader.printVerbose()
			page.pheader.printVerbose()
			page.printFileTrailer()
			page.printDirectorySlots()
		}
		fmt.Printf("\r\n")
	}
}
