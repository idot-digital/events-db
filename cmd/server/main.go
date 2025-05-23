package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/idot-digital/events-db/database"
	pb "github.com/idot-digital/events-db/grpc"
	"github.com/idot-digital/events-db/internal/config"
	"github.com/idot-digital/events-db/internal/handlers"
	"github.com/idot-digital/events-db/internal/middleware"
	"github.com/idot-digital/events-db/internal/server"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

		var s *grpc.Server
		// Configure TLS if certificates are provided
		if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
			log.Info("Starting gRPC server with TLS",
				"cert_file", cfg.TLSCertFile,
				"key_file", cfg.TLSKeyFile)

			creds, err := credentials.NewServerTLSFromFile(cfg.TLSCertFile, cfg.TLSKeyFile)
			if err != nil {
				log.Error("Failed to load gRPC TLS credentials", "error", err)
				os.Exit(1)
			}

			s = grpc.NewServer(
				grpc.Creds(creds),
				grpc.UnaryInterceptor(middleware.AuthInterceptor(cfg.AuthToken)),
				grpc.StreamInterceptor(middleware.StreamAuthInterceptor(cfg.AuthToken)),
			)
		} else {
			log.Info("Starting gRPC server without TLS")
			s = grpc.NewServer(
				grpc.UnaryInterceptor(middleware.AuthInterceptor(cfg.AuthToken)),
				grpc.StreamInterceptor(middleware.StreamAuthInterceptor(cfg.AuthToken)),
			)
		}

		pb.RegisterEventsDBServer(s, grpcHandlers)
		log.Info("gRPC server listening", "address", lis.Addr().String())
		if err := s.Serve(lis); err != nil {
			log.Error("Failed to serve gRPC", "error", err)
			os.Exit(1)
		}
	}()

	// Create a new mux for the REST server
	mux := http.NewServeMux()

	// Add Prometheus metrics endpoint (no auth required)
	mux.Handle("/metrics", promhttp.Handler())

	// Wrap handlers with auth and metrics middleware
	mux.HandleFunc("/events", middleware.Auth(middleware.Metrics(httpHandlers.CreateEventHandler, "create_event"), cfg.AuthToken))
	mux.HandleFunc("/events/get", middleware.Auth(middleware.Metrics(httpHandlers.GetEventByIDHandler, "get_event"), cfg.AuthToken))
	mux.HandleFunc("/events/stream", middleware.Auth(middleware.Metrics(httpHandlers.StreamEventsFromSubjectHandler, "stream_events"), cfg.AuthToken))

	log.Info("REST server listening", "address", fmt.Sprintf(":%d", cfg.RESTPort))

	// Check if TLS certificates are provided
	if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
		log.Info("Starting HTTPS server with TLS",
			"cert_file", cfg.TLSCertFile,
			"key_file", cfg.TLSKeyFile)
		if err := http.ListenAndServeTLS(
			fmt.Sprintf(":%d", cfg.RESTPort),
			cfg.TLSCertFile,
			cfg.TLSKeyFile,
			mux,
		); err != nil {
			log.Error("Failed to serve REST over HTTPS", "error", err)
			os.Exit(1)
		}
	} else {
		log.Info("Starting HTTP server (no TLS)")
		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.RESTPort), mux); err != nil {
			log.Error("Failed to serve REST", "error", err)
			os.Exit(1)
		}
	}
}
