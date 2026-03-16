package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const anthropicAPIVersion = "2023-06-01"

// AnthropicProvider handles the Anthropic Messages API.
type AnthropicProvider struct {
	apiKey string
	models []string
	client *http.Client
}

func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	return &AnthropicProvider{
		apiKey: apiKey,
		models: []string{
			"claude-opus-4",
			"claude-sonnet-4",
			"claude-haiku-4",
			"claude-opus-4-20250514",
			"claude-sonnet-4-20250514",
		},
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (p *AnthropicProvider) Name() string { return "anthropic" }

func (p *AnthropicProvider) SupportsModel(model string) bool {
	for _, m := range p.models {
		if m == model {
			return true
		}
	}
	return false
}

// Anthropic Messages API types

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
	System    string             `json:"system,omitempty"`
	Temperature *float64         `json:"temperature,omitempty"`
	TopP      *float64           `json:"top_p,omitempty"`
	Stop      []string           `json:"stop_sequences,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	ID           string              `json:"id"`
	Type         string              `json:"type"`
	Role         string              `json:"role"`
	Content      []anthropicContent  `json:"content"`
	Model        string              `json:"model"`
	StopReason   string              `json:"stop_reason"`
	Usage        anthropicUsage      `json:"usage"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func (p *AnthropicProvider) ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Convert OpenAI format → Anthropic Messages format
	aReq := anthropicRequest{
		Model:       req.Model,
		MaxTokens:   4096,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        req.Stop,
	}

	if req.MaxTokens != nil {
		aReq.MaxTokens = *req.MaxTokens
	}

	// Extract system message and convert the rest
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			aReq.System = msg.Content
			continue
		}
		aReq.Messages = append(aReq.Messages, anthropicMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Anthropic requires at least one message
	if len(aReq.Messages) == 0 {
		return nil, fmt.Errorf("anthropic requires at least one non-system message")
	}

	body, err := json.Marshal(aReq)
	if err != nil {
		return nil, fmt.Errorf("marshal anthropic request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicAPIVersion)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var aResp anthropicResponse
	if err := json.Unmarshal(respBody, &aResp); err != nil {
		return nil, fmt.Errorf("unmarshal anthropic response: %w", err)
	}

	// Convert Anthropic response → OpenAI format
	content := ""
	for _, c := range aResp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	finishReason := "stop"
	if aResp.StopReason == "max_tokens" {
		finishReason = "length"
	}

	return &ChatResponse{
		ID:      aResp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   aResp.Model,
		Choices: []Choice{
			{
				Index:        0,
				Message:      Message{Role: "assistant", Content: content},
				FinishReason: finishReason,
			},
		},
		Usage: Usage{
			PromptTokens:     aResp.Usage.InputTokens,
			CompletionTokens: aResp.Usage.OutputTokens,
			TotalTokens:      aResp.Usage.InputTokens + aResp.Usage.OutputTokens,
		},
	}, nil
}
