package auth

import (
	"encoding/json"
	"fmt"

	"github.com/zalando/go-keyring"
)

const serviceName = "graphql-cli"

// Store manages credential persistence.
// Primary: OS keyring. Fallback: config file (insecure).
type Store struct{}

func NewStore() *Store {
	return &Store{}
}

// Save stores a credential for the given endpoint.
// Returns (insecureStorageUsed, error).
func (s *Store) Save(endpointName string, cred *Credential) (bool, error) {
	data, err := json.Marshal(cred)
	if err != nil {
		return false, fmt.Errorf("marshal credential: %w", err)
	}

	err = keyring.Set(serviceName, endpointName, string(data))
	if err == nil {
		return false, nil
	}

	// keyring unavailable — fall back to file
	return true, saveToFile(endpointName, data)
}

// Load retrieves a credential for the given endpoint.
func (s *Store) Load(endpointName string) (*Credential, error) {
	// try keyring first
	data, err := keyring.Get(serviceName, endpointName)
	if err == nil {
		var cred Credential
		if err := json.Unmarshal([]byte(data), &cred); err != nil {
			return nil, fmt.Errorf("parse credential: %w", err)
		}

		return &cred, nil
	}

	// try file fallback
	return loadFromFile(endpointName)
}

// Delete removes a credential for the given endpoint.
func (s *Store) Delete(endpointName string) error {
	_ = keyring.Delete(serviceName, endpointName)
	_ = deleteFromFile(endpointName)

	return nil
}
