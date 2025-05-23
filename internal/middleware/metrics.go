package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/idot-digital/events-db/internal/metrics"
)

// MetricsMiddleware wraps an HTTP handler with Prometheus metrics
func Metrics(next http.HandlerFunc, operation string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture the status code
		rw := &responseWriter{ResponseWriter: w}

		next(rw, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		metrics.EventOperationDuration.WithLabelValues(operation).Observe(duration)
		metrics.EventOperations.WithLabelValues(operation, fmt.Sprintf("%d", rw.statusCode)).Inc()
	}
}
