package middleware

import (
	"crypto/tls"
	"log/slog"
	"net/http"
	"time"
)

// Logging wraps an http.Handler and logs requests using slog
func Logging(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		log.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", duration.Milliseconds(),
		)
	})
}

// ConfigureServer sets up a server with proper logging configuration
func ConfigureServer(addr string, handler http.Handler, log *slog.Logger) *http.Server {
	// Create a custom error logger that uses slog
	errorLog := slog.NewLogLogger(log.Handler(), slog.LevelError)

	return &http.Server{
		Addr:     addr,
		Handler:  handler,
		ErrorLog: errorLog,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	}
}
