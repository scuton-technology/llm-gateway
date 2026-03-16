package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/scuton-technology/llm-gateway/internal/admin"
	"github.com/scuton-technology/llm-gateway/internal/middleware"
	"github.com/scuton-technology/llm-gateway/internal/providers"
	"github.com/scuton-technology/llm-gateway/internal/proxy"
	"github.com/scuton-technology/llm-gateway/internal/storage"
)

func main() {
	port := envOr("PORT", "8080")
	dbPath := envOr("DB_PATH", "gateway.db")

	// Initialize storage
	store, err := storage.New(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize provider registry
	registry := providers.NewRegistry()
	registerProviders(registry)

	// Load dashboard HTML
	dashboardPath := envOr("DASHBOARD_PATH", findDashboard())
	dashboardHTML, err := os.ReadFile(dashboardPath)
	if err != nil {
		log.Printf("warning: dashboard not found at %s, /admin will return 404", dashboardPath)
		dashboardHTML = []byte("<html><body><h1>Dashboard not found</h1><p>Set DASHBOARD_PATH env var.</p></body></html>")
	}

	// Create handlers
	router := proxy.NewRouter(registry, store)
	adminHandler := admin.NewHandler(store, dashboardHTML)

	// Setup HTTP mux
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/v1/chat/completions", router.HandleChatCompletion)
	mux.HandleFunc("/health", router.HandleHealth)
	mux.HandleFunc("/api/stats", router.HandleStats)
	mux.HandleFunc("/api/logs", router.HandleLogs)

	// Admin dashboard
	mux.HandleFunc("/admin", adminHandler.ServeDashboard)
	mux.HandleFunc("/api/dashboard", adminHandler.HandleDashboardData)

	// Apply middleware
	handler := middleware.Logging(mux)

	log.Printf("LLM Gateway starting on :%s", port)
	log.Printf("Registered providers: %v", registry.ListProviders())
	log.Printf("POST /v1/chat/completions — proxy endpoint")
	log.Printf("GET  /health              — health check")
	log.Printf("GET  /admin               — dashboard")

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

// findDashboard looks for dashboard.html relative to the executable or CWD.
func findDashboard() string {
	// Try relative to CWD
	candidates := []string{
		"web/dashboard.html",
		"../../web/dashboard.html",
	}

	// Try relative to executable
	if ex, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(ex), "web", "dashboard.html"))
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "web/dashboard.html"
}

func registerProviders(registry *providers.Registry) {
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		registry.Register(providers.NewAnthropicProvider(key))
		log.Println("  + anthropic (claude-opus-4, claude-sonnet-4, claude-haiku-4)")
	}

	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		registry.Register(providers.NewOpenAIProvider(providers.OpenAIConfig{
			Name:    "openai",
			BaseURL: "https://api.openai.com",
			APIKey:  key,
			Models:  []string{"gpt-4o", "gpt-4o-mini", "o1", "o3-mini"},
		}))
		log.Println("  + openai (gpt-4o, gpt-4o-mini, o1, o3-mini)")
	}

	if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
		registry.Register(providers.NewGeminiProvider(key))
		log.Println("  + google (gemini-2.0-flash, gemini-1.5-pro)")
	}

	if key := os.Getenv("GROQ_API_KEY"); key != "" {
		registry.Register(providers.NewOpenAIProvider(providers.OpenAIConfig{
			Name:    "groq",
			BaseURL: "https://api.groq.com/openai",
			APIKey:  key,
			Models:  []string{"llama-3.3-70b-versatile", "mixtral-8x7b-32768"},
		}))
		log.Println("  + groq (llama-3.3-70b, mixtral-8x7b)")
	}

	if key := os.Getenv("MISTRAL_API_KEY"); key != "" {
		registry.Register(providers.NewMistralProvider(key))
		log.Println("  + mistral (mistral-large, mistral-small, codestral)")
	}

	if key := os.Getenv("COHERE_API_KEY"); key != "" {
		registry.Register(providers.NewCohereProvider(key))
		log.Println("  + cohere (command-r-plus, command-r)")
	}

	if key := os.Getenv("XAI_API_KEY"); key != "" {
		registry.Register(providers.NewXAIProvider(key))
		log.Println("  + xai (grok-2, grok-2-mini)")
	}

	if key := os.Getenv("PERPLEXITY_API_KEY"); key != "" {
		registry.Register(providers.NewPerplexityProvider(key))
		log.Println("  + perplexity (sonar-large, sonar-small)")
	}

	if key := os.Getenv("TOGETHER_API_KEY"); key != "" {
		registry.Register(providers.NewOpenAIProvider(providers.OpenAIConfig{
			Name:    "together",
			BaseURL: "https://api.together.xyz",
			APIKey:  key,
			Models:  []string{"meta-llama/Llama-3-70b-chat-hf"},
		}))
		log.Println("  + together (meta-llama/Llama-3-70b)")
	}

	ollamaURL := envOr("OLLAMA_BASE_URL", "http://localhost:11434")
	if os.Getenv("OLLAMA_ENABLED") == "true" {
		registry.Register(providers.NewOpenAIProvider(providers.OpenAIConfig{
			Name:    "ollama",
			BaseURL: ollamaURL,
			Models:  []string{"llama3", "mistral", "codellama", "phi3"},
		}))
		log.Println("  + ollama (local)")
	}

	lmStudioURL := envOr("LMSTUDIO_BASE_URL", "http://localhost:1234")
	if os.Getenv("LMSTUDIO_ENABLED") == "true" {
		registry.Register(providers.NewOpenAIProvider(providers.OpenAIConfig{
			Name:    "lmstudio",
			BaseURL: lmStudioURL,
			Models:  []string{},
		}))
		log.Println("  + lmstudio (local)")
	}

	vllmURL := envOr("VLLM_BASE_URL", "http://localhost:8000")
	if os.Getenv("VLLM_ENABLED") == "true" {
		registry.Register(providers.NewOpenAIProvider(providers.OpenAIConfig{
			Name:    "vllm",
			BaseURL: vllmURL,
			Models:  []string{},
		}))
		log.Println("  + vllm (local)")
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
