package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/idot-digital/events-db/database"
	pb "github.com/idot-digital/events-db/grpc"
	"github.com/idot-digital/events-db/internal/config"
	"github.com/idot-digital/events-db/internal/handlers"
	"github.com/idot-digital/events-db/internal/metrics"
	"github.com/idot-digital/events-db/internal/server"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.New()

	// Initialize logger
	jsonHandler := slog.NewJSONHandler(os.Stderr, nil)
	log := slog.New(jsonHandler)

	// Initialize database connection
	d, err := sql.Open("mysql", cfg.GetDBURI())
	if err != nil {
		log.Error("Failed to open database connection", "error", err)
		os.Exit(1)
	}
	defer d.Close()

	// Read and execute schema.sql
	schemaSQL, err := os.ReadFile("schema.sql")
	if err != nil {
		log.Error("Failed to read schema.sql", "error", err)
		os.Exit(1)
	}

	_, err = d.Exec(string(schemaSQL))
	if err != nil {
		log.Error("Failed to execute schema.sql", "error", err)
		os.Exit(1)
	}

	queries := database.New(d)
	srv := server.New(queries, cfg.EventEmitterBufferLimit, log)
	grpcHandlers := handlers.NewGRPCHandlers(srv)
	httpHandlers := handlers.NewHTTPHandlers(srv)

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
		if err != nil {
			log.Error("Failed to listen for gRPC", "error", err)
			os.Exit(1)
		}
		s := grpc.NewServer()
		pb.RegisterEventsDBServer(s, grpcHandlers)
		log.Info("gRPC server listening", "address", lis.Addr().String())
		if err := s.Serve(lis); err != nil {
			log.Error("Failed to serve gRPC", "error", err)
			os.Exit(1)
		}
	}()

	// Create a new mux for the REST server
	mux := http.NewServeMux()

	// Add Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Wrap handlers with metrics middleware
	mux.HandleFunc("/events", metricsMiddleware(httpHandlers.CreateEventHandler, "create_event"))
	mux.HandleFunc("/events/get", metricsMiddleware(httpHandlers.GetEventByIDHandler, "get_event"))
	mux.HandleFunc("/events/stream", metricsMiddleware(httpHandlers.StreamEventsFromSubjectHandler, "stream_events"))

	log.Info("REST server listening", "address", fmt.Sprintf(":%d", cfg.RESTPort))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.RESTPort), mux); err != nil {
		log.Error("Failed to serve REST", "error", err)
		os.Exit(1)
	}
}

// metricsMiddleware wraps an HTTP handler with Prometheus metrics
func metricsMiddleware(next http.HandlerFunc, operation string) http.HandlerFunc {
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

// responseWriter is a custom response writer that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriter) CloseNotify() <-chan bool {
	if cn, ok := rw.ResponseWriter.(http.CloseNotifier); ok {
		return cn.CloseNotify()
	}
	return nil
}

func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
}
