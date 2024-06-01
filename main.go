package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	aliasFile()
	getSubscriptions()
}

func aliasFile() []byte {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error finding home directory:", err)
		return nil
	}

	path := filepath.Join(home, ".azure_aliases")
	aliases, err := os.ReadFile(path)

	if err != nil {
		createAliasesFile(path)
	}
	return aliases
}

func createAliasesFile(filePath string) {
	content := []byte("# Azure Aliases\n")
	err := os.WriteFile(filePath, content, 0644)
	if err != nil {
		log.Fatalln("Error creating file:", err)
	}
}

func checkCli() string {
	path, err := exec.LookPath("az")
	if err != nil {
		log.Fatalf("Unable to locate %v", err)
	}
	return path
}

func getSubscriptions() []byte {
	var cmd *exec.Cmd
	path := checkCli()

	fmt.Println("CLI", path)

	cmd = exec.Command(path, "account", "list", "--query", "'[].{id:id, name:name}", "-o", "tsv")
	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("Could not execute command %v", err)
	}
	fmt.Println(string(out))

	return out

}
