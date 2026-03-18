package ai

import (
	"fmt"
	"sort"
	"sync"

	"gitlens/internal/config"
)

var (
	mu       sync.RWMutex
	registry = map[string]ProviderFactory{}
)

// Register adds a provider factory. Call from init() in each provider file.
func Register(name string, factory ProviderFactory) {
	mu.Lock()
	defer mu.Unlock()
	registry[name] = factory
}

// New creates a Provider from the resolved config.
func New(cfg *config.Config) (Provider, error) {
	mu.RLock()
	factory, ok := registry[cfg.Provider]
	mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown AI provider %q (registered: %v)", cfg.Provider, registeredNames())
	}
	return factory(cfg.APIKey, cfg.Model)
}

func registeredNames() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(registry))
	for k := range registry {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
