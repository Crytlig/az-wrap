package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

func main() {
	alias := flag.String("alias", "", "Set a subscription alias by <subscriptionId>:<alias>")
	flag.Parse()
	if *alias != "" {
		parts := strings.SplitN(*alias, ":", 2)
		if len(parts) != 2 {
			log.Fatalln("Invalid alias format. Use <subscriptionId>:<alias>")
		}
		saveAliasFile(parts[0], parts[1])
		os.Exit(0)
	}

	aliases := subscriptionAliases()
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Index", "Alias", "Name", "ID")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for _, s := range aliases {
		id := s.ID
		if s.Selected {
			id = color.New(color.BgBlue, color.FgWhite).Sprint(id)
		}
		tbl.AddRow(s.Index, s.Alias, s.Name, id)
	}
	tbl.Print()
	var selection string
	color.New(color.FgGreen).Print("\nEnter Index, Alias, Name or ID to select: ")
	fmt.Scan(&selection)
	selection = strings.ToLower(selection)

	for _, s := range aliases {
		if selection == strconv.Itoa(s.Index) ||
			selection == strings.ToLower(s.Alias) ||
			selection == strings.ToLower(s.Name) ||
			selection == strings.ToLower(s.ID) {
			fmt.Printf("Selected %s with ID %s\n", s.Name, s.ID)
			setSubscription(s.ID)
			os.Exit(0)
		}
	}
	fmt.Println("No subscriptions found")
}
