package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/scuton-technology/llm-gateway/internal/middleware"
	"github.com/scuton-technology/llm-gateway/internal/storage"
)

type Handler struct {
	store         *storage.Store
	dashboardHTML []byte
}

func NewHandler(store *storage.Store, dashboardHTML []byte) *Handler {
	return &Handler{store: store, dashboardHTML: dashboardHTML}
}

// ServeDashboard serves the dashboard HTML page.
func (h *Handler) ServeDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(h.dashboardHTML)
}

// HandleDashboardData returns JSON data for the dashboard.
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
		"stats":       stats,
		"recent_logs": logs,
		"cost_table":  middleware.CostPerMillionTokens,
	})
}
