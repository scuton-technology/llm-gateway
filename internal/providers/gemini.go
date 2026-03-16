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

// GeminiProvider handles Google Gemini API.
type GeminiProvider struct {
	apiKey string
	models []string
	client *http.Client
}

func NewGeminiProvider(apiKey string) *GeminiProvider {
	return &GeminiProvider{
		apiKey: apiKey,
		models: []string{"gemini-2.0-flash", "gemini-1.5-pro"},
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (p *GeminiProvider) Name() string { return "google" }

func (p *GeminiProvider) SupportsModel(model string) bool {
	for _, m := range p.models {
		if m == model {
			return true
		}
	}
	return false
}

// Gemini API types

type geminiRequest struct {
	Contents         []geminiContent         `json:"contents"`
	SystemInstruction *geminiContent          `json:"systemInstruction,omitempty"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
	UsageMetadata *geminiUsage   `json:"usageMetadata"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

type geminiUsage struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

func (p *GeminiProvider) ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	gReq := geminiRequest{}

	// Build generation config
	gc := &geminiGenerationConfig{
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxOutputTokens: req.MaxTokens,
		StopSequences:   req.Stop,
	}
	gReq.GenerationConfig = gc

	// Convert messages
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			gReq.SystemInstruction = &geminiContent{
				Parts: []geminiPart{{Text: msg.Content}},
			}
			continue
		}

		role := msg.Role
		if role == "assistant" {
			role = "model"
		}

		gReq.Contents = append(gReq.Contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: msg.Content}},
		})
	}

	body, err := json.Marshal(gReq)
	if err != nil {
		return nil, fmt.Errorf("marshal gemini request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		req.Model, p.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var gResp geminiResponse
	if err := json.Unmarshal(respBody, &gResp); err != nil {
		return nil, fmt.Errorf("unmarshal gemini response: %w", err)
	}

	content := ""
	finishReason := "stop"
	if len(gResp.Candidates) > 0 {
		for _, part := range gResp.Candidates[0].Content.Parts {
			content += part.Text
		}
		if gResp.Candidates[0].FinishReason == "MAX_TOKENS" {
			finishReason = "length"
		}
	}

	usage := Usage{}
	if gResp.UsageMetadata != nil {
		usage = Usage{
			PromptTokens:     gResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: gResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      gResp.UsageMetadata.TotalTokenCount,
		}
	}

	return &ChatResponse{
		ID:      fmt.Sprintf("gemini-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []Choice{
			{
				Index:        0,
				Message:      Message{Role: "assistant", Content: content},
				FinishReason: finishReason,
			},
		},
		Usage: usage,
	}, nil
}
