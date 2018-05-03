package main

import (
	"fmt"
	"os"
	"spf13/cobra"
)

type searchOptions struct {
	file string
}

func newSearchCommand() *cobra.Command {
	var options searchOptions
	c := &cobra.Command{
		Use:   "search",
		Short: "search the primary key in innodb ibd file",
		Long:  "search the primary key in innodb ibd file",
		Run: func(cmd *cobra.Command, args []string) {
			doSearch(cmd, &options)
		},
	}

	c.Flags().StringVarP(&options.file, "file", "f", "", "innodb table space file path")

	return c
}

func doSearch(cmd *cobra.Command, options *searchOptions) {
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

	_, err = parseInnodbDataFile(f, &parsePageOptions{
		parsePageTypeFlag: parsePageInode,
	})
	if nil != err {
		fmt.Println("Parse innodb data file error ", err)
	}
}
