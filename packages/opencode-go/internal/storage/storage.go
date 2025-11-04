package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/opencode/opencode-go/internal/config"
)

type Storage struct {
	dir string
}

// New creates a new storage instance
// If path is empty, uses the default location in XDG_DATA_HOME/opencode/storage
func New(path string) (*Storage, error) {
	// Use default path if not provided
	if path == "" {
		path = filepath.Join(config.GetDataDir(), "storage")
	}

	// Ensure directory exists
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}

	return &Storage{dir: path}, nil
}

func (s *Storage) Close() error {
	return nil
}

// Read reads a file from the hierarchical path
func (s *Storage) Read(key []string) ([]byte, error) {
	target := filepath.Join(s.dir, filepath.Join(key...)) + ".json"
	data, err := os.ReadFile(target)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return data, nil
}

// Write writes a file to the hierarchical path
func (s *Storage) Write(key []string, value []byte) error {
	target := filepath.Join(s.dir, filepath.Join(key...)) + ".json"

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	return os.WriteFile(target, value, 0644)
}

// Delete removes a file from the hierarchical path
func (s *Storage) Delete(key []string) error {
	target := filepath.Join(s.dir, filepath.Join(key...)) + ".json"
	err := os.Remove(target)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// ReadJSON reads and unmarshals a JSON file
func (s *Storage) ReadJSON(key []string, v interface{}) error {
	data, err := s.Read(key)
	if err != nil {
		return err
	}
	if data == nil {
		return nil
	}
	return json.Unmarshal(data, v)
}

// WriteJSON marshals and writes a JSON file
func (s *Storage) WriteJSON(key []string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return s.Write(key, data)
}

// List returns all items under a prefix path
// For example, List(["session", "project123"]) returns all session IDs for project123
func (s *Storage) List(prefix []string) ([][]string, error) {
	basePath := filepath.Join(s.dir, filepath.Join(prefix...))

	// Check if directory exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return [][]string{}, nil
	}

	var results [][]string
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".json") {
			// Get relative path from basePath
			relPath, err := filepath.Rel(basePath, path)
			if err != nil {
				return err
			}

			// Remove .json extension
			relPath = strings.TrimSuffix(relPath, ".json")

			// Split into parts
			parts := strings.Split(filepath.ToSlash(relPath), "/")

			// Build full key path
			fullKey := append(prefix, parts...)
			results = append(results, fullKey)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// Legacy API compatibility methods - these delegate to the new path-based API
func (s *Storage) Get(bucket, key string) ([]byte, error) {
	return s.Read([]string{bucket, key})
}

func (s *Storage) Set(bucket, key string, value []byte) error {
	return s.Write([]string{bucket, key}, value)
}

func (s *Storage) GetJSON(bucket, key string, v interface{}) error {
	return s.ReadJSON([]string{bucket, key}, v)
}

func (s *Storage) SetJSON(bucket, key string, v interface{}) error {
	return s.WriteJSON([]string{bucket, key}, v)
}
