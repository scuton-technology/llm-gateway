package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ChatRequest represents the OpenAI-compatible chat completion request format.
// This is the universal API format used across all providers.
type ChatRequest struct {
	Model             string            `json:"model"`
	Messages          []Message         `json:"messages"`
	Temperature       *float64          `json:"temperature,omitempty"`
	MaxTokens         *int              `json:"max_tokens,omitempty"`
	TopP              *float64          `json:"top_p,omitempty"`
	Stream            bool              `json:"stream,omitempty"`
	Stop              []string          `json:"stop,omitempty"`
	N                 int               `json:"n,omitempty"`
	Tools             []json.RawMessage `json:"tools,omitempty"`
	ToolChoice        any               `json:"tool_choice,omitempty"`
	ResponseFormat    any               `json:"response_format,omitempty"`
	StreamOptions     any               `json:"stream_options,omitempty"`
	User              string            `json:"user,omitempty"`
	PresencePenalty   *float64          `json:"presence_penalty,omitempty"`
	FrequencyPenalty  *float64          `json:"frequency_penalty,omitempty"`
	ParallelToolCalls *bool             `json:"parallel_tool_calls,omitempty"`
}

type Message struct {
	Role       string         `json:"role"`
	Content    MessageContent `json:"content"`
	Name       string         `json:"name,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall     `json:"tool_calls,omitempty"`
	Refusal    string         `json:"refusal,omitempty"`
}

func (m Message) TextContent() (string, error) {
	return m.Content.Text()
}

type MessageContent struct {
	raw json.RawMessage
}

func NewTextContent(text string) MessageContent {
	raw, _ := json.Marshal(text)
	return MessageContent{raw: raw}
}

func (c *MessageContent) UnmarshalJSON(data []byte) error {
	c.raw = append(c.raw[:0], data...)
	return nil
}

func (c MessageContent) MarshalJSON() ([]byte, error) {
	if len(c.raw) == 0 {
		return []byte("null"), nil
	}
	return c.raw, nil
}

func (c MessageContent) Text() (string, error) {
	if len(c.raw) == 0 || string(c.raw) == "null" {
		return "", nil
	}

	var text string
	if err := json.Unmarshal(c.raw, &text); err == nil {
		return text, nil
	}

	var parts []ContentPart
	if err := json.Unmarshal(c.raw, &parts); err == nil {
		var builder strings.Builder
		for _, part := range parts {
			switch part.Type {
			case "", "text", "input_text":
				builder.WriteString(part.textValue())
			default:
				return "", fmt.Errorf("unsupported content part type %q", part.Type)
			}
		}
		return builder.String(), nil
	}

	var part ContentPart
	if err := json.Unmarshal(c.raw, &part); err == nil && (part.Type != "" || part.Text != "" || part.Refusal != "") {
		switch part.Type {
		case "", "text", "input_text":
			return part.textValue(), nil
		default:
			return "", fmt.Errorf("unsupported content part type %q", part.Type)
		}
	}

	return "", fmt.Errorf("unsupported message content format")
}

type ContentPart struct {
	Type     string `json:"type,omitempty"`
	Text     string `json:"text,omitempty"`
	Refusal  string `json:"refusal,omitempty"`
	ImageURL any    `json:"image_url,omitempty"`
}

func (p ContentPart) textValue() string {
	if p.Text != "" {
		return p.Text
	}
	return p.Refusal
}

// ChatResponse represents the OpenAI-compatible chat completion response format.
type ChatResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
	SystemFingerprint string   `json:"system_fingerprint,omitempty"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	Logprobs     any     `json:"logprobs,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id,omitempty"`
	Type     string       `json:"type,omitempty"`
	Function ToolFunction `json:"function,omitempty"`
}

type ToolFunction struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Provider is the interface that all LLM providers must implement.
type Provider interface {
	// Name returns the provider identifier (e.g., "openai", "anthropic").
	Name() string

	// ChatCompletion sends a chat completion request and returns the response.
	ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error)

	// SupportsModel returns true if this provider can handle the given model.
	SupportsModel(model string) bool
}

// ErrorResponse represents an API error in OpenAI format.
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}
