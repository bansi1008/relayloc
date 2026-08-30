package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Credentials struct {
	AgentID    string `json:"agent_id"`
	AgentName  string `json:"agent_name,omitempty"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

func SaveCredentials(filePath string, creds Credentials) error {
	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create config dir: %w", err)
		}
	}

	credsMap := make(map[string]Credentials)
	if data, err := os.ReadFile(filePath); err == nil {
		var existingMap map[string]Credentials
		if err := json.Unmarshal(data, &existingMap); err == nil {
			credsMap = existingMap
		} else {
			var single Credentials
			if err := json.Unmarshal(data, &single); err == nil && single.AgentName != "" {
				credsMap[strings.ToLower(single.AgentName)] = single
			}
		}
	}

	key := strings.ToLower(creds.AgentName)
	if key == "" {
		key = strings.ToLower(creds.AgentID)
	}
	credsMap[key] = creds

	data, err := json.MarshalIndent(credsMap, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	return os.WriteFile(filePath, data, 0600)
}

func GetCredentialsByName(filePath string, name string) (*Credentials, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read credentials: %w", err)
	}

	target := strings.ToLower(strings.TrimSpace(name))

	var credsMap map[string]Credentials
	if err := json.Unmarshal(data, &credsMap); err == nil {
		if cred, ok := credsMap[target]; ok {
			return &cred, nil
		}
		for _, cred := range credsMap {
			if strings.EqualFold(cred.AgentName, name) || strings.EqualFold(cred.AgentID, name) {
				return &cred, nil
			}
		}
	}

	var single Credentials
	if err := json.Unmarshal(data, &single); err == nil {
		if strings.EqualFold(single.AgentName, name) || strings.EqualFold(single.AgentID, name) {
			return &single, nil
		}
	}

	return nil, fmt.Errorf("credentials for agent %q not found in local store", name)
}
