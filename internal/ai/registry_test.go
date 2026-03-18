package ai_test

import (
	"testing"

	"gitlens/internal/ai"
	"gitlens/internal/config"
)

func TestRegistryUnknownProvider(t *testing.T) {
	cfg := &config.Config{Provider: "nonexistent-xyz"}
	_, err := ai.New(cfg)
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

func TestRegistryKnownProvider(t *testing.T) {
	ai.Register("testprovider", func(apiKey, model string) (ai.Provider, error) {
		return &mockProvider{name: "testprovider"}, nil
	})
	cfg := &config.Config{Provider: "testprovider", APIKey: "key", Model: "m"}
	p, err := ai.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "testprovider" {
		t.Errorf("name: got %q, want testprovider", p.Name())
	}
}
