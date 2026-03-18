package ai

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func init() {
	Register("claude", func(apiKey, model string) (Provider, error) {
		if apiKey == "" {
			return nil, fmt.Errorf("claude: api_key is required")
		}
		if model == "" {
			model = "claude-opus-4-6"
		}
		client := anthropic.NewClient(option.WithAPIKey(apiKey))
		return &claudeProvider{
			client: client,
			model:  model,
		}, nil
	})
}

type claudeProvider struct {
	client anthropic.Client
	model  string
}

func (c *claudeProvider) Name() string { return "claude" }

func (c *claudeProvider) Complete(ctx context.Context, prompt string) (string, error) {
	msg, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("claude complete: %w", err)
	}
	if len(msg.Content) == 0 {
		return "", fmt.Errorf("claude: empty response")
	}
	return msg.Content[0].Text, nil
}

func (c *claudeProvider) Stream(ctx context.Context, prompt string) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk)
	go func() {
		defer close(ch)
		stream := c.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
			Model:     anthropic.Model(c.model),
			MaxTokens: 1024,
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
			},
		})
		for stream.Next() {
			event := stream.Current()
			// content_block_delta events carry text in Delta.Text
			if event.Type == "content_block_delta" && event.Delta.Text != "" {
				select {
				case ch <- StreamChunk{Text: event.Delta.Text}:
				case <-ctx.Done():
					return
				}
			}
		}
		if err := stream.Err(); err != nil {
			select {
			case ch <- StreamChunk{Err: fmt.Errorf("claude stream: %w", err)}:
			case <-ctx.Done():
			}
		}
	}()
	return ch, nil
}
