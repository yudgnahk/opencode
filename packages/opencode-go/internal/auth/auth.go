package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/opencode/opencode-go/internal/config"
)

// AuthType represents the type of authentication
type AuthType string

const (
	TypeOAuth     AuthType = "oauth"
	TypeAPI       AuthType = "api"
	TypeWellKnown AuthType = "wellknown"
)

// OAuth represents OAuth authentication with refresh/access tokens
type OAuth struct {
	Type    AuthType `json:"type"`
	Refresh string   `json:"refresh"`
	Access  string   `json:"access"`
	Expires int64    `json:"expires"`
}

// API represents simple API key authentication
type API struct {
	Type AuthType `json:"type"`
	Key  string   `json:"key"`
}

// WellKnown represents well-known authentication with key and token
type WellKnown struct {
	Type  AuthType `json:"type"`
	Key   string   `json:"key"`
	Token string   `json:"token"`
}

// Info is a discriminated union type that can be any auth type
type Info interface {
	GetType() AuthType
}

func (o OAuth) GetType() AuthType     { return o.Type }
func (a API) GetType() AuthType       { return a.Type }
func (w WellKnown) GetType() AuthType { return w.Type }

// mutex protects concurrent access to auth.json
var mu sync.Mutex

// authFilePathFunc is the function used to get the auth file path
// Can be overridden for testing
var authFilePathFunc = getDefaultAuthFilePath

// getDefaultAuthFilePath returns the path to auth.json in the data directory
func getDefaultAuthFilePath() string {
	return filepath.Join(config.GetDataDir(), "auth.json")
}

// getAuthFilePath returns the current auth file path
func getAuthFilePath() string {
	return authFilePathFunc()
}

// Get retrieves authentication info for a specific provider
// Returns nil if provider not found or file doesn't exist
func Get(providerID string) (Info, error) {
	mu.Lock()
	defer mu.Unlock()

	all, err := readAuthFile()
	if err != nil {
		return nil, err
	}

	data, exists := all[providerID]
	if !exists {
		return nil, nil
	}

	return parseAuthInfo(data)
}

// All retrieves all authentication records
// Returns empty map if file doesn't exist
func All() (map[string]Info, error) {
	mu.Lock()
	defer mu.Unlock()

	rawData, err := readAuthFile()
	if err != nil {
		return nil, err
	}

	result := make(map[string]Info, len(rawData))
	for key, data := range rawData {
		info, err := parseAuthInfo(data)
		if err != nil {
			return nil, err
		}
		if info != nil {
			result[key] = info
		}
	}

	return result, nil
}

// Set saves or updates authentication info for a provider
func Set(key string, info Info) error {
	mu.Lock()
	defer mu.Unlock()

	all, err := readAuthFile()
	if err != nil {
		return err
	}

	all[key] = info
	return writeAuthFile(all)
}

// Remove deletes authentication info for a provider
func Remove(key string) error {
	mu.Lock()
	defer mu.Unlock()

	all, err := readAuthFile()
	if err != nil {
		return err
	}

	delete(all, key)
	return writeAuthFile(all)
}

// readAuthFile reads and parses auth.json
// Returns empty map if file doesn't exist
func readAuthFile() (map[string]interface{}, error) {
	filepath := getAuthFilePath()

	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// writeAuthFile writes auth data to auth.json with 0o600 permissions
func writeAuthFile(data map[string]interface{}) error {
	filepath := getAuthFilePath()

	// Ensure directory exists
	dir := filepath[:len(filepath)-len("/auth.json")]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Marshal with indentation like TypeScript (2 spaces)
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Write file with restricted permissions
	if err := os.WriteFile(filepath, jsonData, 0o600); err != nil {
		return err
	}

	// Ensure permissions are set (some systems need explicit chmod)
	return os.Chmod(filepath, 0o600)
}

// parseAuthInfo parses raw JSON data into the appropriate auth type
func parseAuthInfo(data interface{}) (Info, error) {
	// Convert to JSON and back to properly handle the type
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// First, determine the type
	var typeCheck struct {
		Type AuthType `json:"type"`
	}
	if err := json.Unmarshal(jsonData, &typeCheck); err != nil {
		return nil, err
	}

	// Parse based on type
	switch typeCheck.Type {
	case TypeOAuth:
		var oauth OAuth
		if err := json.Unmarshal(jsonData, &oauth); err != nil {
			return nil, err
		}
		return oauth, nil
	case TypeAPI:
		var api API
		if err := json.Unmarshal(jsonData, &api); err != nil {
			return nil, err
		}
		return api, nil
	case TypeWellKnown:
		var wellknown WellKnown
		if err := json.Unmarshal(jsonData, &wellknown); err != nil {
			return nil, err
		}
		return wellknown, nil
	default:
		return nil, nil
	}
}
