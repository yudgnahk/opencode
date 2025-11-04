package modelsdev

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	ModelsDevURL = "https://models.dev/api.json"
	UserAgent    = "opencode-go/1.0"
)

// Model represents a model from models.dev
type Model struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	ReleaseDate  string                 `json:"release_date"`
	Attachment   bool                   `json:"attachment"`
	Reasoning    bool                   `json:"reasoning"`
	Temperature  bool                   `json:"temperature"`
	ToolCall     bool                   `json:"tool_call"`
	Cost         ModelCost              `json:"cost"`
	Limit        ModelLimit             `json:"limit"`
	Modalities   *ModelModalities       `json:"modalities,omitempty"`
	Experimental bool                   `json:"experimental,omitempty"`
	Status       string                 `json:"status,omitempty"` // "alpha", "beta", "deprecated"
	Options      map[string]interface{} `json:"options"`
	Headers      map[string]string      `json:"headers,omitempty"`
	Provider     *ModelProvider         `json:"provider,omitempty"`
}

type ModelCost struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cache_read,omitempty"`
	CacheWrite float64 `json:"cache_write,omitempty"`
}

type ModelLimit struct {
	Context int `json:"context"`
	Output  int `json:"output"`
}

type ModelModalities struct {
	Input  []string `json:"input"`
	Output []string `json:"output"`
}

type ModelProvider struct {
	NPM string `json:"npm"`
}

// Provider represents a provider from models.dev
type Provider struct {
	ID     string           `json:"id"`
	Name   string           `json:"name"`
	Env    []string         `json:"env"`
	API    string           `json:"api,omitempty"`
	NPM    string           `json:"npm,omitempty"`
	Models map[string]Model `json:"models"`
}

// Database holds all providers from models.dev
type Database map[string]Provider

// Client handles fetching and caching models.dev data
type Client struct {
	cacheDir   string
	httpClient *http.Client
}

// NewClient creates a new models.dev client
func NewClient(cacheDir string) *Client {
	return &Client{
		cacheDir: cacheDir,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Get retrieves the models database, using cache if available
func (c *Client) Get() (Database, error) {
	// Try to load from cache first
	cached, err := c.loadCache()
	if err == nil {
		return cached, nil
	}

	// Fallback to fetching from network
	return c.Fetch()
}

// Fetch downloads the latest models database from models.dev
func (c *Client) Fetch() (Database, error) {
	req, err := http.NewRequest("GET", ModelsDevURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models.dev: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("models.dev returned status %d", resp.StatusCode)
	}

	var db Database
	if err := json.NewDecoder(resp.Body).Decode(&db); err != nil {
		return nil, fmt.Errorf("failed to decode models.dev response: %w", err)
	}

	// Save to cache
	if err := c.saveCache(db); err != nil {
		// Log but don't fail - caching is optional
		fmt.Fprintf(os.Stderr, "Warning: failed to save cache: %v\n", err)
	}

	return db, nil
}

// loadCache attempts to load the cached models database
func (c *Client) loadCache() (Database, error) {
	cachePath := c.getCachePath()

	file, err := os.Open(cachePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var db Database
	if err := json.NewDecoder(file).Decode(&db); err != nil {
		return nil, err
	}

	return db, nil
}

// saveCache saves the models database to cache
func (c *Client) saveCache(db Database) error {
	cachePath := c.getCachePath()

	// Ensure cache directory exists
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return err
	}

	file, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(db)
}

// getCachePath returns the path to the cache file
func (c *Client) getCachePath() string {
	return filepath.Join(c.cacheDir, "models.json")
}

// RefreshInBackground starts a background goroutine to refresh the cache periodically
func (c *Client) RefreshInBackground(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if _, err := c.Fetch(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to refresh models.dev cache: %v\n", err)
			}
		}
	}()
}
