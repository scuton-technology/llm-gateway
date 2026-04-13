package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logging middleware logs HTTP requests to stdout.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)

		log.Printf("%s %s %d %s %s",
			r.Method, r.URL.Path, ww.status,
			time.Since(start).Round(time.Millisecond), ClientIP(r))
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
