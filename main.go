package main

import (
	"fmt"
	"spf13/cobra"
)

func main() {
	var cmdEntry = &cobra.Command{Use: "innoisp"}
	cmdEntry.AddCommand(newOverviewCommand())
	cmdEntry.AddCommand(newDslotsCommand())
	if err := cmdEntry.Execute(); nil != err {
		fmt.Println("command execute error ", err)
	}
}
