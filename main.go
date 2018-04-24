package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	flagFile string
)

func main() {
	flag.StringVar(&flagFile, "file", "", "input innodb ibd file")
	flag.Parse()

	if "" == flagFile {
		fmt.Println("No input file specified")
		return
	}

	f, err := os.Open(flagFile)
	if nil != err {
		fmt.Println("Open file error ", err)
		return
	}
	defer f.Close()

	pages, err := parseInnodbDataFile(f)
	if nil != err {
		fmt.Println("Parse innodb data file error ", err)
	}
	printPages(pages)
}