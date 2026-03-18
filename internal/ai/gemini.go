package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func init() {
	Register("gemini", func(apiKey, model string) (Provider, error) {
		if apiKey == "" {
			return nil, fmt.Errorf("gemini: api_key is required")
		}
		if model == "" {
			model = "gemini-2.0-flash"
		}
		client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
		if err != nil {
			return nil, fmt.Errorf("gemini client: %w", err)
		}
		return &geminiProvider{
			client: client,
			model:  model,
		}, nil
	})
}

type geminiProvider struct {
	client *genai.Client
	model  string
}

func (g *geminiProvider) Name() string { return "gemini" }

func (g *geminiProvider) Close() error {
	return g.client.Close()
}

func (g *geminiProvider) Complete(ctx context.Context, prompt string) (string, error) {
	m := g.client.GenerativeModel(g.model)
	resp, err := m.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("gemini complete: %w", err)
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini: empty response")
	}
	var parts []string
	for _, p := range resp.Candidates[0].Content.Parts {
		parts = append(parts, fmt.Sprintf("%s", p))
	}
	return strings.Join(parts, ""), nil
}

func (g *geminiProvider) Stream(ctx context.Context, prompt string) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk)
	go func() {
		defer close(ch)
		m := g.client.GenerativeModel(g.model)
		iter := m.GenerateContentStream(ctx, genai.Text(prompt))
		for {
			resp, err := iter.Next()
			if err == iterator.Done {
				return
			}
			if err != nil {
				select {
				case ch <- StreamChunk{Err: fmt.Errorf("gemini stream: %w", err)}:
				case <-ctx.Done():
				}
				return
			}
			for _, cand := range resp.Candidates {
				for _, p := range cand.Content.Parts {
					select {
					case ch <- StreamChunk{Text: fmt.Sprintf("%s", p)}:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return ch, nil
}
