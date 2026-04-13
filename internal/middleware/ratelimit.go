package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a simple token bucket rate limiter per IP.
type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*bucket
	rate    int // requests per window
	window  time.Duration
}

type bucket struct {
	tokens    int
	lastReset time.Time
}

func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*bucket),
		rate:    rate,
		window:  window,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := ClientIP(r)

		rl.mu.Lock()
		b, exists := rl.clients[ip]
		if !exists {
			b = &bucket{tokens: rl.rate, lastReset: time.Now()}
			rl.clients[ip] = b
		}

		// Reset bucket if window has passed
		if time.Since(b.lastReset) > rl.window {
			b.tokens = rl.rate
			b.lastReset = time.Now()
		}

		if b.tokens <= 0 {
			rl.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"message":"rate limit exceeded","type":"error","code":"rate_limit"}}`))
			return
		}

		b.tokens--
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
