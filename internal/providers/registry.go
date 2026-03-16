package providers

import (
	"fmt"
	"strings"
	"sync"
)

// Registry maps model names to providers.
type Registry struct {
	mu        sync.RWMutex
	providers []Provider
	// prefixMap maps model prefixes to provider names for fast lookup.
	// e.g. "claude-" → "anthropic", "gpt-" → "openai"
	prefixMap map[string]string
	// exactMap maps exact model names to provider names.
	// e.g. "o1" → "openai", "o3-mini" → "openai"
	exactMap map[string]string
}

func NewRegistry() *Registry {
	return &Registry{
		prefixMap: map[string]string{
			"claude-":  "anthropic",
			"gpt-":     "openai",
			"gemini-":  "google",
			"mistral-": "mistral",
			"command-": "cohere",
			"grok-":    "xai",
			"sonar-":   "perplexity",
			"mixtral-": "groq",
			"llama":    "groq",
		},
		exactMap: map[string]string{
			"o1":              "openai",
			"o3-mini":         "openai",
			"codestral-latest": "mistral",
		},
	}
}

// Register adds a provider to the registry.
func (r *Registry) Register(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers = append(r.providers, p)
}

// Resolve finds the appropriate provider for a given model name.
func (r *Registry) Resolve(model string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providerName := r.resolveProviderName(model)
	if providerName == "" {
		return nil, fmt.Errorf("no provider found for model %q", model)
	}

	for _, p := range r.providers {
		if p.Name() == providerName {
			return p, nil
		}
	}

	// Fallback: check all providers for explicit model support
	for _, p := range r.providers {
		if p.SupportsModel(model) {
			return p, nil
		}
	}

	return nil, fmt.Errorf("provider %q resolved for model %q but not registered", providerName, model)
}

func (r *Registry) resolveProviderName(model string) string {
	// Check exact matches first
	if name, ok := r.exactMap[model]; ok {
		return name
	}

	// Check prefix matches
	lower := strings.ToLower(model)
	for prefix, name := range r.prefixMap {
		if strings.HasPrefix(lower, prefix) {
			return name
		}
	}

	// Check if model contains a slash (e.g., "meta-llama/Llama-3-70b")
	// These are typically Together AI or HuggingFace-style model IDs
	if strings.Contains(model, "/") {
		return "together"
	}

	return ""
}

// ListProviders returns all registered provider names.
func (r *Registry) ListProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, len(r.providers))
	for i, p := range r.providers {
		names[i] = p.Name()
	}
	return names
}

// ReplaceWith swaps the internal state of this registry with another.
// This enables hot-reloading providers without replacing the pointer.
func (r *Registry) ReplaceWith(other *Registry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	r.providers = other.providers
	r.prefixMap = other.prefixMap
	r.exactMap = other.exactMap
}

// ListModels returns all supported models across all providers.
func (r *Registry) ListModels() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var models []string
	// Collect from exact map
	for model := range r.exactMap {
		models = append(models, model)
	}
	return models
}
