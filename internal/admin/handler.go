package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/scuton-technology/llm-gateway/internal/middleware"
	"github.com/scuton-technology/llm-gateway/internal/storage"
)

type Handler struct {
	store *storage.Store
}

func NewHandler(store *storage.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) HandleDashboardData(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.GetStats(24 * time.Hour)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logs, err := h.store.GetRecentLogs(50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"stats":        stats,
		"recent_logs":  logs,
		"cost_table":   middleware.CostPerMillionTokens,
	})
}
