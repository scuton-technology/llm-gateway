package providers

import (
	"encoding/json"
	"testing"
)

func TestMessageContentTextFromParts(t *testing.T) {
	var msg Message
	if err := json.Unmarshal([]byte(`{"role":"user","content":[{"type":"text","text":"hello "},{"type":"text","text":"world"}]}`), &msg); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	text, err := msg.TextContent()
	if err != nil {
		t.Fatalf("TextContent() error = %v", err)
	}
	if text != "hello world" {
		t.Fatalf("TextContent() = %q, want %q", text, "hello world")
	}
}

func TestMessageContentRejectsUnsupportedParts(t *testing.T) {
	var msg Message
	if err := json.Unmarshal([]byte(`{"role":"user","content":[{"type":"image_url","image_url":{"url":"https://example.com/a.png"}}]}`), &msg); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if _, err := msg.TextContent(); err == nil {
		t.Fatalf("TextContent() error = nil, want unsupported content error")
	}
}

func TestRegistryPrefersExplicitProviderModelSupport(t *testing.T) {
	registry := NewRegistry()
	registry.Register(NewOpenAIProvider(OpenAIConfig{
		Name:    "groq",
		BaseURL: "https://api.groq.com/openai",
		Models:  []string{"llama-3.3-70b-versatile"},
	}))
	registry.Register(NewOpenAIProvider(OpenAIConfig{
		Name:    "ollama",
		BaseURL: "http://localhost:11434",
		Models:  []string{"llama3"},
	}))

	provider, err := registry.Resolve("llama3")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if provider.Name() != "ollama" {
		t.Fatalf("Resolve() provider = %q, want %q", provider.Name(), "ollama")
	}
}
