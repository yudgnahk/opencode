package modelsdev

import (
	"os"
	"path/filepath"
	"sync"
)

// Registry provides access to models.dev data with caching
type Registry struct {
	client *Client
	db     Database
	mu     sync.RWMutex
}

// NewRegistry creates a new models.dev registry
func NewRegistry(cacheDir string) *Registry {
	return &Registry{
		client: NewClient(cacheDir),
	}
}

// Load loads the models database from cache or network
func (r *Registry) Load() error {
	db, err := r.client.Get()
	if err != nil {
		return err
	}

	r.mu.Lock()
	r.db = db
	r.mu.Unlock()

	return nil
}

// GetProvider returns provider information by ID
func (r *Registry) GetProvider(providerID string) (*Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.db[providerID]
	return &provider, ok
}

// GetModels returns all model IDs for a given provider
func (r *Registry) GetModels(providerID string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.db[providerID]
	if !ok {
		return nil
	}

	models := make([]string, 0, len(provider.Models))
	for modelID := range provider.Models {
		models = append(models, modelID)
	}

	return models
}

// GetModel returns model information by provider ID and model ID
func (r *Registry) GetModel(providerID, modelID string) (*Model, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.db[providerID]
	if !ok {
		return nil, false
	}

	model, ok := provider.Models[modelID]
	return &model, ok
}

// GetDefaultCacheDir returns the default cache directory
func GetDefaultCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".cache/opencode"
	}
	return filepath.Join(home, ".cache", "opencode")
}
