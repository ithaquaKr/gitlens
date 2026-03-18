package ai_test

import (
	"context"

	"gitlens/internal/ai"
)

type mockProvider struct{ name string }

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) Complete(_ context.Context, prompt string) (string, error) {
	n := len(prompt)
	if n > 20 {
		n = 20
	}
	return "mock response for: " + prompt[:n], nil
}

func (m *mockProvider) Stream(_ context.Context, prompt string) (<-chan ai.StreamChunk, error) {
	ch := make(chan ai.StreamChunk, 1)
	ch <- ai.StreamChunk{Text: "mock"}
	close(ch)
	return ch, nil
}
