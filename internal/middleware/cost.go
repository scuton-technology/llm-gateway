package middleware

// CostPerMillionTokens defines pricing per model (input/output per 1M tokens in USD).
var CostPerMillionTokens = map[string][2]float64{
	// Anthropic
	"claude-opus-4":   {15.0, 75.0},
	"claude-sonnet-4": {3.0, 15.0},
	"claude-haiku-4":  {0.25, 1.25},

	// OpenAI
	"gpt-4o":      {2.50, 10.0},
	"gpt-4o-mini": {0.15, 0.60},
	"o1":          {15.0, 60.0},
	"o3-mini":     {1.10, 4.40},

	// Google
	"gemini-2.0-flash": {0.10, 0.40},
	"gemini-1.5-pro":   {1.25, 5.00},

	// Groq
	"llama-3.3-70b-versatile": {0.59, 0.79},
	"mixtral-8x7b-32768":      {0.24, 0.24},

	// Mistral
	"mistral-large-latest": {2.0, 6.0},
	"mistral-small-latest": {0.2, 0.6},
	"codestral-latest":     {0.3, 0.9},

	// Cohere
	"command-r-plus": {2.50, 10.0},
	"command-r":      {0.15, 0.60},

	// xAI
	"grok-2":      {2.0, 10.0},
	"grok-2-mini": {0.20, 1.0},

	// Perplexity
	"sonar-large": {1.0, 1.0},
	"sonar-small": {0.20, 0.20},
}

// EstimateCost calculates the estimated cost in USD for a request.
func EstimateCost(model string, inputTokens, outputTokens int) float64 {
	prices, ok := CostPerMillionTokens[model]
	if !ok {
		// Prefix match: "claude-sonnet-4-20250514" → "claude-sonnet-4"
		for key, p := range CostPerMillionTokens {
			if len(model) > len(key) && model[:len(key)] == key {
				prices = p
				ok = true
				break
			}
		}
	}
	if !ok {
		return 0
	}

	inputCost := float64(inputTokens) / 1_000_000 * prices[0]
	outputCost := float64(outputTokens) / 1_000_000 * prices[1]
	return inputCost + outputCost
}
