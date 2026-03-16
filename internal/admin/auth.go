package admin

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/scuton-technology/llm-gateway/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost     = 12
	sessionDuration = 24 * time.Hour
	cookieName     = "llm_gateway_session"
)

// AuthMiddleware protects admin routes.
func AuthMiddleware(store *storage.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Always allow these paths without auth
		if path == "/admin/login" || path == "/admin/setup" ||
			path == "/health" || path == "/v1/chat/completions" ||
			strings.HasPrefix(path, "/api/stats") ||
			path == "/api/logs" {
			next.ServeHTTP(w, r)
			return
		}

		// API endpoints used by settings/dashboard need auth too
		// But only admin/* and api/settings* and api/dashboard need protection
		needsAuth := strings.HasPrefix(path, "/admin") ||
			strings.HasPrefix(path, "/api/settings") ||
			path == "/api/dashboard"

		if !needsAuth {
			next.ServeHTTP(w, r)
			return
		}

		// If no password is set, redirect to setup
		if !store.HasAdminPassword() {
			if path != "/admin/setup" {
				http.Redirect(w, r, "/admin/setup", http.StatusFound)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		// Check session cookie
		cookie, err := r.Cookie(cookieName)
		if err != nil || !store.ValidateSession(cookie.Value) {
			// For API calls, return 401
			if strings.HasPrefix(path, "/api/") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]any{"error": "unauthorized"})
				return
			}
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AuthHandler handles login, logout, and setup.
type AuthHandler struct {
	store     *storage.Store
	loginHTML []byte
	setupHTML []byte
}

func NewAuthHandler(store *storage.Store, loginHTML, setupHTML []byte) *AuthHandler {
	return &AuthHandler{
		store:     store,
		loginHTML: loginHTML,
		setupHTML: setupHTML,
	}
}

// ServeLogin serves the login page.
func (ah *AuthHandler) ServeLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// If already logged in, redirect to admin
		if cookie, err := r.Cookie(cookieName); err == nil && ah.store.ValidateSession(cookie.Value) {
			http.Redirect(w, r, "/admin", http.StatusFound)
			return
		}
		// If no password set, redirect to setup
		if !ah.store.HasAdminPassword() {
			http.Redirect(w, r, "/admin/setup", http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(ah.loginHTML)
		return
	}

	if r.Method == http.MethodPost {
		ah.handleLogin(w, r)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (ah *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	ip := clientIPFromRequest(r)

	// Check brute force lockout
	if ah.store.IsIPLocked(ip) {
		remaining := ah.store.GetLockoutRemaining(ip)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]any{
			"error":       "too many failed attempts, try again later",
			"locked_for":  remaining,
		})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		Password string `json:"password"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	hash, err := ah.store.GetAdminPasswordHash()
	if err != nil || hash == "" {
		http.Error(w, "no password configured", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		ah.store.RecordLoginAttempt(ip, false)
		log.Printf("Failed login attempt from %s", ip)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{"error": "invalid password"})
		return
	}

	// Success — create session
	ah.store.RecordLoginAttempt(ip, true)
	token, err := ah.store.CreateSession(sessionDuration)
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	// Clean up expired sessions periodically
	ah.store.CleanExpiredSessions()

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionDuration.Seconds()),
	})

	log.Printf("Successful login from %s", ip)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "redirect": "/admin"})
}

// ServeSetup serves the initial setup page.
func (ah *AuthHandler) ServeSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// If password already set, redirect to login
		if ah.store.HasAdminPassword() {
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(ah.setupHTML)
		return
	}

	if r.Method == http.MethodPost {
		ah.handleSetup(w, r)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (ah *AuthHandler) handleSetup(w http.ResponseWriter, r *http.Request) {
	// Only allow setup if no password exists
	if ah.store.HasAdminPassword() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]any{"error": "password already set"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		Password        string `json:"password"`
		PasswordConfirm string `json:"password_confirm"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 6 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{"error": "password must be at least 6 characters"})
		return
	}

	if req.Password != req.PasswordConfirm {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{"error": "passwords do not match"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	if err := ah.store.SetAdminPassword(string(hash)); err != nil {
		http.Error(w, "failed to save password", http.StatusInternalServerError)
		return
	}

	log.Printf("Admin password set up successfully")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "redirect": "/admin/login"})
}

// HandleLogout destroys the session and redirects to login.
func (ah *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(cookieName); err == nil {
		ah.store.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/admin/login", http.StatusFound)
}

func clientIPFromRequest(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.TrimSpace(strings.Split(ip, ",")[0])
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	// Strip port from RemoteAddr (e.g. "127.0.0.1:12345" → "127.0.0.1")
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		// Handle IPv6 like [::1]:port
		if strings.Contains(addr, "]") {
			if bracketIdx := strings.LastIndex(addr, "]"); bracketIdx != -1 {
				return addr[1:bracketIdx] // strip [ and ]
			}
		}
		return addr[:idx]
	}
	return addr
}
