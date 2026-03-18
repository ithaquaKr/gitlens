package ai

import (
	"fmt"
	"sort"

	"gitlens/internal/config"
)

var registry = map[string]ProviderFactory{}

// Register adds a provider factory. Call from init() in each provider file.
func Register(name string, factory ProviderFactory) {
	registry[name] = factory
}

// New creates a Provider from the resolved config.
func New(cfg *config.Config) (Provider, error) {
	factory, ok := registry[cfg.Provider]
	if !ok {
		return nil, fmt.Errorf("unknown AI provider %q (registered: %v)", cfg.Provider, registeredNames())
	}
	return factory(cfg.APIKey, cfg.Model)
}

func registeredNames() []string {
	names := make([]string, 0, len(registry))
	for k := range registry {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
