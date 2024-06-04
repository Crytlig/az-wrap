package main

import (
	"context"
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
	ctx := context.Background()
	alias := parseFlags()
	if err := handleAliasFlag(alias); err != nil {
		log.Fatalln(err)
	}

	aliases, err := subscriptionAliases()
	if err != nil {
		log.Fatalln(err)
	}

	displayAliases(aliases)

	selection := promptUserForSelection()
	if err := selectSubscription(ctx, aliases, selection); err != nil {
		fmt.Println("No subscriptions found")
	}
}

func parseFlags() string {
	alias := flag.String("alias", "", "Set a subscription alias by <subscriptionId>:<alias>")
	flag.Parse()
	return *alias
}

func handleAliasFlag(alias string) error {
	if alias != "" {
		parts := strings.SplitN(alias, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid alias format. Use <subscriptionId>:<alias>")
		}
		if err := saveAliasFile(parts[0], parts[1]); err != nil {
			return fmt.Errorf("error saving alias: %w", err)
		}
		os.Exit(0)
	}
	return nil
}

func displayAliases(aliases []subscriptionAlias) {
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
}

func promptUserForSelection() string {
	var selection string
	color.New(color.FgGreen).Print("\nEnter Index, Alias, Name or ID to select: ")
	fmt.Scan(&selection)
	return strings.ToLower(selection)
}

func selectSubscription(ctx context.Context, aliases []subscriptionAlias, selection string) error {
	for _, s := range aliases {
		if selection == strconv.Itoa(s.Index) ||
			selection == strings.ToLower(s.Alias) ||
			selection == strings.ToLower(s.Name) ||
			selection == strings.ToLower(s.ID) {
			fmt.Printf("Selected %s with ID %s\n", s.Name, s.ID)
			if err := setSubscription(ctx, s.ID); err != nil {
				return err
			}
			os.Exit(0)
		}
	}
	return fmt.Errorf("subscription not found")
}
