package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func SaveAliasFile(subscriptionId, alias string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error finding home directory:", err)
		return
	}
	aliasFile, err := checkAliasFile(homeDir)
	if aliasFile == "" {
		createAliasFile(aliasFile)
	}

	f, err := os.OpenFile(aliasFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening alias file:", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(subscriptionId + ":" + alias + "\n"); err != nil {
		fmt.Println("Error writing to alias file", err)
	} else {
		fmt.Printf("Alias '%s' added for subscription ID '%s'. \n", alias, subscriptionId)
	}
}

// Stupid simple implementation. Maybe the full json object might be better
// For now this works subscriptionID:alias
func LoadAliases() map[string]string {
	aliases := make(map[string]string)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return aliases
	}

	file, err := checkAliasFile(homeDir)
	if err != nil {
		createAliasFile(file)
		return aliases
	}

	f, err := os.Open(file)

	if err != nil {
		fmt.Println(err)
		return aliases
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
	return aliases
}

func checkAliasFile(homeDir string) (string, error) {
	azureDir := fmt.Sprintf("%s/.azure", homeDir)
	if _, err := os.Stat(azureDir); os.IsNotExist(err) {
		return "", err
	}

	aliasPath := fmt.Sprintf("%s/aliases", azureDir)
	return aliasPath, nil
}

func createAliasFile(file string) {
	f, err := os.Create(file)
	if err != nil {
		fmt.Println("Error creating alias file:", err)
		return
	}
	defer f.Close()
}
