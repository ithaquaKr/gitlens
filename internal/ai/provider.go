package ai

import "context"

// Provider is the interface all AI backends must implement.
type Provider interface {
	Complete(ctx context.Context, prompt string) (string, error)
	// Stream sends text chunks; a chunk with Err != nil signals failure; channel close = done.
	// Note: lumen uses non-streaming; Stream() is a gitlens enhancement.
	Stream(ctx context.Context, prompt string) (<-chan StreamChunk, error)
	Name() string
}

// StreamChunk is one token from a streaming response.
type StreamChunk struct {
	Text string
	Err  error
}

// ProviderFactory constructs a Provider from credentials.
type ProviderFactory func(apiKey, model string) (Provider, error)
