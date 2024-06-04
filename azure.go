package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type loadedSubscriptions struct {
	Name     string `json:"name"`
	ID       string `json:"id"`
	Selected bool   `json:"isDefault"`
}

type subscriptionAlias struct {
	Name     string
	ID       string
	Index    int
	Alias    string
	Selected bool
}

// subscriptionAliases retrieves subscription aliases.
func subscriptionAliases() ([]subscriptionAlias, error) {
	subs, err := subscriptions()
	if err != nil {
		return nil, err
	}
	aliases, err := loadAliases()
	if err != nil {
		return nil, err
	}

	var subscriptionAliases []subscriptionAlias
	for i, sub := range subs {
		alias := aliases[sub.ID]
		if alias == "" {
			alias = "(no alias)"
		}
		subscriptionAliases = append(subscriptionAliases, subscriptionAlias{
			Name:     sub.Name,
			ID:       sub.ID,
			Index:    i + 1,
			Alias:    alias,
			Selected: sub.Selected,
		})
	}
	return subscriptionAliases, nil
}

// setSubscription sets the Azure subscription using given ID.
func setSubscription(ctx context.Context, ID string) error {
	path, err := azureCLIPath()
	if err != nil {
		return fmt.Errorf("unable to use the Azure CLI for setting subscription: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "account", "set", "--subscription", ID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("unable to set subscription: %w", err)
	}

	return nil
}

func subscriptions() ([]loadedSubscriptions, error) {
	// Prefer file over Azure CLI. It loads ~ 90% quicker
	subs, err := getSubscriptionsFromFile()
	if err != nil {
		subs, err = getSubscriptionsWithCLI(context.Background())
	}

	if err != nil {
		return nil, err
	}

	return subs, nil
}

func getSubscriptionsFromFile() ([]loadedSubscriptions, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("unable to find a home directory: %w", err)
	}

	azureProfilePath := filepath.Join(homeDir, ".azure", "azureProfile.json")
	if _, err := os.Stat(azureProfilePath); err != nil {
		return nil, fmt.Errorf("unable to locate: %w", err)
	}

	file, err := os.ReadFile(azureProfilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read azureProfile.json: %w", err)
	}

	// The file is encoded with UTF-8 BOM for some reason
	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf")) // Or []byte{239, 187, 191}

	// azureProfile.json contains additional fields
	// so only look in the subscriptions array
	var s struct {
		Subscriptions []loadedSubscriptions `json:"subscriptions"`
	}
	if err := json.Unmarshal(file, &s); err != nil {
		return nil, fmt.Errorf("unable to parse azureProfile.json: %w", err)
	}

	return s.Subscriptions, err
}

func getSubscriptionsWithCLI(ctx context.Context) ([]loadedSubscriptions, error) {
	path, err := azureCLIPath()
	if err != nil {
		return nil, fmt.Errorf("unable to use the Azure CLI for getting subscriptions: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "account", "list")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("command could not run: %w", err)
	}

	var subscriptions []loadedSubscriptions
	if err := json.Unmarshal(out, &subscriptions); err != nil {
		return nil, fmt.Errorf("there was an error unmarshalling Azure CLI accounts: %w", err)
	}

	if len(subscriptions) == 0 {
		return nil, fmt.Errorf("unable to fetch any of your subscriptions with Azure CLI. Please login using 'az login'")
	}

	return subscriptions, err
}

func azureCLIPath() (string, error) {
	path, err := exec.LookPath("az")
	if err != nil {
		return "", fmt.Errorf("unable to locate az CLI: %w", err)
	}
	return path, nil
}
