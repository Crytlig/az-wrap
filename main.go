package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

func main() {
	aliases := SubscriptionAliases()
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
			SetSubscription(s.ID)
			os.Exit(0)
		}
	}
	fmt.Println("No subscriptions found")
}
