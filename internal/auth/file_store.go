package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func credFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "graphql-cli", "credentials.json")
}

func loadAllFromFile() (map[string]*Credential, error) {
	data, err := os.ReadFile(credFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*Credential), nil
		}

		return nil, err
	}

	var creds map[string]*Credential
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}

	return creds, nil
}

func saveAllToFile(creds map[string]*Credential) error {
	path := credFilePath()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

func saveToFile(endpointName string, credData []byte) error {
	creds, err := loadAllFromFile()
	if err != nil {
		return fmt.Errorf("load credentials file: %w", err)
	}

	var cred Credential
	if err := json.Unmarshal(credData, &cred); err != nil {
		return err
	}

	creds[endpointName] = &cred

	return saveAllToFile(creds)
}

func loadFromFile(endpointName string) (*Credential, error) {
	creds, err := loadAllFromFile()
	if err != nil {
		return nil, err
	}

	cred, ok := creds[endpointName]
	if !ok {
		return nil, nil
	}

	return cred, nil
}

func deleteFromFile(endpointName string) error {
	creds, err := loadAllFromFile()
	if err != nil {
		return err
	}

	delete(creds, endpointName)

	return saveAllToFile(creds)
}
