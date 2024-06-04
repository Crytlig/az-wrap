package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// saveAliasFile saves the alias to a file
func saveAliasFile(subscriptionId, alias string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error finding home directory: %w", err)
	}

	aliasFile, err := checkAliasFile(homeDir)
	if err != nil {
		if err := createAliasFile(aliasFile); err != nil {
			return fmt.Errorf("error creating alias file: %w", err)
		}
	}

	if err := writeAliasToFile(aliasFile, subscriptionId, alias); err != nil {
		return fmt.Errorf("error writing to alias file: %w", err)
	}

	fmt.Printf("Alias '%s' added for subscription ID '%s'.\n", alias, subscriptionId)
	return nil
}

func writeAliasToFile(aliasFile, subscriptionId, alias string) error {
	f, err := os.OpenFile(aliasFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening alias file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(subscriptionId + ":" + alias + "\n"); err != nil {
		return fmt.Errorf("error writing to alias file: %w", err)
	}
	return nil
}

// loadAliases loads aliases from the alias file
func loadAliases() (map[string]string, error) {
	aliases := make(map[string]string)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return aliases, fmt.Errorf("error finding home directory: %w", err)
	}

	file, err := checkAliasFile(homeDir)
	if err != nil {
		return aliases, nil // No aliases file is not a hard error
	}

	f, err := os.Open(file)
	if err != nil {
		return aliases, fmt.Errorf("error opening alias file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			aliases[parts[0]] = parts[1]
		}
	}
	return aliases, scanner.Err()
}

// checkAliasFile checks if the alias file exists and returns its path
func checkAliasFile(homeDir string) (string, error) {
	aliasPath := filepath.Join(homeDir, ".azure", "aliases")
	if _, err := os.Stat(aliasPath); os.IsNotExist(err) {
		return aliasPath, err
	}
	return aliasPath, nil
}

// createAliasFile creates the alias file
func createAliasFile(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("error creating alias file: %w", err)
	}
	defer f.Close()
	return nil
}
