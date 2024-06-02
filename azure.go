package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
)

type loadedSubscriptions struct {
	Name     string `json:"name"`
	ID       string `json:"id"`
	Selected bool   `json:"isDefault"`
}

type SubscriptionAlias struct {
	Name     string
	ID       string
	Index    int
	Alias    string
	Selected bool
}

func SubscriptionAliases() []SubscriptionAlias {
	subs := subscriptions()
	aliases := LoadAliases()

	var subscriptionAliases []SubscriptionAlias
	for i, sub := range subs {

		alias := aliases[sub.ID]
		if alias == "" {
			alias = "(no alias)"
		}
		subscriptionAliases = append(subscriptionAliases, SubscriptionAlias{
			Name:     sub.Name,
			ID:       sub.ID,
			Index:    i + 1,
			Alias:    alias,
			Selected: sub.Selected,
		})
	}
	return subscriptionAliases
}

func SetSubscription(ID string) {
	path, err := azureCLIPath()
	if err != nil {
		log.Fatalf("Unable to use the Azure CLI for setting subscription: %v", err)
	}

	cmd := exec.Command(path, "account", "set", "--subscription", ID)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Unable to set subscription: %v", err)
	}
}

func subscriptions() []loadedSubscriptions {
	// Prefer file over Azure CLI. It loads ~ 90% quicker
	subs, err := getSubscriptionsFromFile()
	if err != nil {
		subs, err = getSubscriptionsWithCLI()
	}

	if err != nil {
		log.Fatal(err)
	}

	return subs
}

func getSubscriptionsFromFile() ([]loadedSubscriptions, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("unable to find a home directory: %v", err)
	}

	azureProfilePath := homeDir + "/.azure/azureProfile.json"
	if _, err := os.Stat(azureProfilePath); err != nil {
		return nil, fmt.Errorf("unable to locate: %v", err)
	}

	file, err := os.ReadFile(azureProfilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read azureProfile.json: %v", err)
	}

	// The file is encoded with UTF-8 BOM for some reason
	// https://stackoverflow.com/questions/31398044/got-error-invalid-character-%C3%AF-looking-for-beginning-of-value-from-json-unmar
	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf")) // Or []byte{239, 187, 191}

	// azureProfile.json contains additional fields
	// so only look in the subscriptions array
	var s struct {
		Subscriptions []loadedSubscriptions `json:"subscriptions"`
	}
	if err := json.Unmarshal(file, &s); err != nil {
		return nil, fmt.Errorf("unable to parse azureProfile.json: %v", err)
	}

	return s.Subscriptions, err
}

func getSubscriptionsWithCLI() ([]loadedSubscriptions, error) {
	path, err := azureCLIPath()
	if err != nil {
		log.Fatalf("Unable to use the Azure CLI for getting subscriptions: %v", err)
	}
	cmd := exec.Command(path, "account", "list")
	out, err := cmd.Output()

	if err != nil {
		fmt.Println("Command could not run")
	}

	var subscriptions []loadedSubscriptions

	if err := json.Unmarshal(out, &subscriptions); err != nil {
		return nil, fmt.Errorf("There was an error unmarshalling Azure CLI accounts: %v", err)
	}

	if len(subscriptions) == 0 {
		log.Fatal("Unable to fetch any of your subscriptions with Azure CLI. Please login using 'az login'")
	}
	return subscriptions, err
}

func azureCLIPath() (string, error) {
	path, err := exec.LookPath("az")
	if err != nil {
		log.Fatalf("Unable to locate az CLI: %v", err)
	}
	return path, err
}
