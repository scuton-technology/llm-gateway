package providers

// Together AI uses the OpenAI-compatible format.
// Use NewOpenAIProvider with Together AI config:
//
//   NewOpenAIProvider(OpenAIConfig{
//       Name:    "together",
//       BaseURL: "https://api.together.xyz",
//       APIKey:  togetherAPIKey,
//       Models:  []string{"meta-llama/Llama-3-70b-chat-hf", ...},
//   })
//
// No separate adapter needed — the OpenAI adapter handles it.
