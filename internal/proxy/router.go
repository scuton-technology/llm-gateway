package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/scuton-technology/llm-gateway/internal/providers"
	"github.com/scuton-technology/llm-gateway/internal/storage"
)

type Router struct {
	registry *providers.Registry
	store    *storage.Store
}

func NewRouter(registry *providers.Registry, store *storage.Store) *Router {
	return &Router{
		registry: registry,
		store:    store,
	}
}

func (rt *Router) HandleChatCompletion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", "invalid_request")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body", "invalid_request")
		return
	}
	defer r.Body.Close()

	var req providers.ChatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error(), "invalid_request")
		return
	}

	if req.Model == "" {
		writeError(w, http.StatusBadRequest, "model field is required", "invalid_request")
		return
	}

	if len(req.Messages) == 0 {
		writeError(w, http.StatusBadRequest, "messages field is required", "invalid_request")
		return
	}

	// Resolve provider
	provider, err := rt.registry.Resolve(req.Model)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "model_not_found")
		return
	}

	// Execute request
	start := time.Now()
	resp, err := provider.ChatCompletion(r.Context(), req)
	latency := time.Since(start)

	// Prepare log entry
	logEntry := storage.RequestLog{
		Model:     req.Model,
		Provider:  provider.Name(),
		LatencyMs: latency.Milliseconds(),
		ClientIP:  clientIP(r),
	}

	if err != nil {
		logEntry.StatusCode = http.StatusBadGateway
		logEntry.ErrorMessage = err.Error()
		rt.logRequest(logEntry)
		writeError(w, http.StatusBadGateway, "provider error: "+err.Error(), "provider_error")
		return
	}

	// Success
	logEntry.StatusCode = http.StatusOK
	logEntry.PromptTokens = resp.Usage.PromptTokens
	logEntry.CompletionTokens = resp.Usage.CompletionTokens
	logEntry.TotalTokens = resp.Usage.TotalTokens
	rt.logRequest(logEntry)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-LLM-Provider", provider.Name())
	w.Header().Set("X-LLM-Latency-Ms", fmt.Sprintf("%d", latency.Milliseconds()))
	json.NewEncoder(w).Encode(resp)
}

func (rt *Router) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":    "ok",
		"providers": rt.registry.ListProviders(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (rt *Router) HandleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := rt.store.GetStats(24 * time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get stats", "internal_error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (rt *Router) HandleLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := rt.store.GetRecentLogs(100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get logs", "internal_error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func (rt *Router) logRequest(entry storage.RequestLog) {
	if err := rt.store.LogRequest(entry); err != nil {
		log.Printf("failed to log request: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(providers.ErrorResponse{
		Error: struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		}{
			Message: message,
			Type:    "error",
			Code:    code,
		},
	})
}

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}
