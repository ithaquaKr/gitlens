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
		return &geminiProvider{apiKey: apiKey, model: model}, nil
	})
}

type geminiProvider struct {
	apiKey string
	model  string
}

func (g *geminiProvider) Name() string { return "gemini" }

func (g *geminiProvider) Complete(ctx context.Context, prompt string) (string, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(g.apiKey))
	if err != nil {
		return "", fmt.Errorf("gemini client: %w", err)
	}
	defer client.Close()
	m := client.GenerativeModel(g.model)
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
		client, err := genai.NewClient(ctx, option.WithAPIKey(g.apiKey))
		if err != nil {
			ch <- StreamChunk{Err: fmt.Errorf("gemini client: %w", err)}
			return
		}
		defer client.Close()
		m := client.GenerativeModel(g.model)
		iter := m.GenerateContentStream(ctx, genai.Text(prompt))
		for {
			resp, err := iter.Next()
			if err == iterator.Done {
				return
			}
			if err != nil {
				ch <- StreamChunk{Err: fmt.Errorf("gemini stream: %w", err)}
				return
			}
			for _, cand := range resp.Candidates {
				for _, p := range cand.Content.Parts {
					ch <- StreamChunk{Text: fmt.Sprintf("%s", p)}
				}
			}
		}
	}()
	return ch, nil
}
