package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
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

type config struct {
	homeDir      string
	azureDir     string
	aliasFile    string
	azureProfile string
}

func newConfig() (*config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("unable to find home directory: %w", err)
	}

	azureDir := filepath.Join(homeDir, ".azure")
	return &config{
		homeDir:      homeDir,
		azureDir:     azureDir,
		azureProfile: filepath.Join(azureDir, "azureProfile.json"),
		aliasFile:    filepath.Join(azureDir, "aliases"),
	}, nil
}

// subscriptionAliases retrieves subscription aliases.
func (c *config) subscriptionAliases() ([]subscriptionAlias, error) {
	subs, err := c.subscriptions()
	if err != nil {
		return nil, err
	}
	aliases, err := c.aliases()
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

// setSubscription sets the Azure subscription using a given ID.
func (c *config) setSubscription(ctx context.Context, ID string) error {
	path, err := c.azureCLIPath()
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

// subscriptions retrieves the list of subscriptions.
func (c *config) subscriptions() ([]loadedSubscriptions, error) {
	// Prefer file over Azure CLI. It loads ~ 90% quicker.
	subs, err := c.getSubscriptionsFromFile()
	if err != nil {
		subs, err = c.getSubscriptionsWithCLI(context.Background())
	}

	if err != nil {
		return nil, err
	}

	return subs, nil
}

// getSubscriptionsFromFile reads subscriptions from the local profile file.
func (c *config) getSubscriptionsFromFile() ([]loadedSubscriptions, error) {
	if _, err := os.Stat(c.azureProfile); err != nil {
		return nil, fmt.Errorf("unable to locate: %w", err)
	}

	file, err := os.ReadFile(c.azureProfile)
	if err != nil {
		return nil, fmt.Errorf("unable to read %s: %w", c.azureProfile, err)
	}

	// The file is encoded with UTF-8 BOM for some reason.
	// https://stackoverflow.com/questions/31398044/got-error-invalid-character-%C3%AF-looking-for-beginning-of-value-from-json-unmar
	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))

	var s struct {
		Subscriptions []loadedSubscriptions `json:"subscriptions"`
	}
	if err := json.Unmarshal(file, &s); err != nil {
		return nil, fmt.Errorf("unable to parse azureProfile.json: %w", err)
	}
	if len(s.Subscriptions) == 0 {
		// %w to get the error chain. Output will be from getSubscriptionsWithCLI path and ask the user to login using az login
		return nil, fmt.Errorf("%w", err)
	}

	return s.Subscriptions, err
}

// getSubscriptionsWithCLI retrieves subscriptions using the Azure CLI.
func (c *config) getSubscriptionsWithCLI(ctx context.Context) ([]loadedSubscriptions, error) {
	path, err := c.azureCLIPath()
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

	// This is an actual error where we want to quit
	loginErr := color.New(color.FgYellow)
	if len(subscriptions) == 0 {
		loginErr.Println("Unable to fetch any of your subscriptions with Azure CLI. Please login using 'az login'")
		os.Exit(0)
	}

	return subscriptions, err
}

func (c *config) azureCLIPath() (string, error) {
	path, err := exec.LookPath("az")
	if err != nil {
		return "", fmt.Errorf("unable to locate az CLI: %w", err)
	}
	return path, nil
}

func (c *config) saveAliasFile(subscriptionId, alias string) error {
	aliasFile, err := c.checkAliasFile()
	if err != nil {
		if err := c.createAliasFile(aliasFile); err != nil {
			return fmt.Errorf("error creating alias file: %w", err)
		}
	}

	if err := c.writeAliasToFile(aliasFile, subscriptionId, alias); err != nil {
		return fmt.Errorf("error writing to alias file: %w", err)
	}

	fmt.Printf("Alias '%s' added for subscription ID '%s'.\n", alias, subscriptionId)
	return nil
}

func (c *config) writeAliasToFile(aliasFile, subscriptionId, alias string) error {
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

// aliases loads aliases from the alias file.
func (c *config) aliases() (map[string]string, error) {
	aliases := make(map[string]string)
	file, err := c.checkAliasFile()
	if err != nil {
		return aliases, nil // No aliases file is not a hard error.
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

func (c *config) checkAliasFile() (string, error) {
	aliasPath := filepath.Join(c.homeDir, ".azure", "aliases")
	if _, err := os.Stat(aliasPath); os.IsNotExist(err) {
		return aliasPath, err
	}
	return aliasPath, nil
}

func (c *config) createAliasFile(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("error creating alias file: %w", err)
	}
	defer f.Close()
	return nil
}
