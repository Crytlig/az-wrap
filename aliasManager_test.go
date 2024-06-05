package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TODO: Fix the config when testing,
// it should not use the default paths
// Implement overriding maybe?
func TestNewConfig(t *testing.T) {
	_, err := newConfig()
	if err != nil {
		t.Fatalf("Failed to create new config: %v", err)
	}
}

func TestCreateAliasFile(t *testing.T) {
	c, err := newConfig()
	if err != nil {
		t.Fatalf("Failed to create new config: %v", err)
	}

	// Use a temporary file for testing
	tempDir := t.TempDir()
	c.homeDir = tempDir
	aliasFile := filepath.Join(tempDir, ".azure", "aliases")
	os.MkdirAll(filepath.Dir(aliasFile), 0755)

	err = c.createAliasFile(aliasFile)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(aliasFile); os.IsNotExist(err) {
		t.Fatalf("Alias file was not created: %v", err)
	}
}

func TestSaveAliasFile(t *testing.T) {
	c, err := newConfig()
	if err != nil {
		t.Fatalf("Failed to create new config: %v", err)
	}

	// Use a temporary file for testing
	tempDir := t.TempDir()
	c.homeDir = tempDir
	aliasFile := filepath.Join(tempDir, ".azure", "aliases")
	os.MkdirAll(filepath.Dir(aliasFile), 0755)

	subscriptionId := "test-subscription-id"
	alias := "test-alias"

	err = c.saveAliasFile(subscriptionId, alias)
	if err != nil {
		t.Fatalf("Failed to save alias file: %v", err)
	}

	// Check that the alias was written correctly
	content, err := os.ReadFile(aliasFile)
	if err != nil {
		t.Fatalf("Failed to read alias file: %v", err)
	}

	expectedContent := subscriptionId + ":" + alias + "\n"
	if string(content) != expectedContent {
		t.Fatalf("Alias file content mismatch. Got: %s, Expected: %s", string(content), expectedContent)
	}
}

func TestSetSubscription(t *testing.T) {
	c, err := newConfig()
	if err != nil {
		t.Fatalf("Failed to create new config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// This is a dummy test; normally you would mock external dependencies.
	subscriptionID := "some-subscription-id"
	err = c.setSubscription(ctx, subscriptionID)
	if err == nil {
		t.Fatalf("Expected error when setting subscription with dummy ID, got none")
	}
}

func TestGetSubscriptionsFromFile(t *testing.T) {
	c, err := newConfig()
	if err != nil {
		t.Fatalf("Failed to create new config: %v", err)
	}

	// Use a temporary file for testing
	tempDir := t.TempDir()
	c.homeDir = tempDir
	azureProfilePath := filepath.Join(tempDir, ".azure", "azureProfile.json")

	// Create dummy azureProfile.json
	profileContent := `{
		"subscriptions": [
			{
				"name": "Test Subscription",
				"id": "test-id",
				"isDefault": true
			}
		]
	}`
	os.MkdirAll(filepath.Dir(azureProfilePath), 0755)
	err = os.WriteFile(azureProfilePath, []byte(profileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write dummy azureProfile.json: %v", err)
	}

	subs, err := c.getSubscriptionsFromFile()
	if err != nil {
		t.Fatalf("Failed to get subscriptions from file: %v", err)
	}

	if len(subs) != 1 || subs[0].ID != "test-id" {
		t.Fatalf("Subscription content mismatch. Got: %+v", subs)
	}
}

func TestAliases(t *testing.T) {
	c, err := newConfig()
	if err != nil {
		t.Fatalf("Failed to create new config: %v", err)
	}

	// Use a temporary file for testing
	tempDir := t.TempDir()
	c.homeDir = tempDir
	aliasFile := filepath.Join(tempDir, ".azure", "aliases")

	// Create dummy alias file
	aliasContent := "test-subscription-id:test-alias\n"
	os.MkdirAll(filepath.Dir(aliasFile), 0755)
	err = os.WriteFile(aliasFile, []byte(aliasContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write dummy alias file: %v", err)
	}

	aliases, err := c.aliases()
	if err != nil {
		t.Fatalf("Failed to load aliases: %v", err)
	}

	expectedAlias := "test-alias"
	if aliases["test-subscription-id"] != expectedAlias {
		t.Fatalf("Alias content mismatch. Got: %s, Expected: %s", aliases["test-subscription-id"], expectedAlias)
	}
}
