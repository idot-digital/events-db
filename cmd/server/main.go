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
	"github.com/idot-digital/events-db/internal/server"
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
		panic(err)
	}
	defer d.Close()

	// Read and execute schema.sql
	schemaSQL, err := os.ReadFile("schema.sql")
	if err != nil {
		panic(fmt.Sprintf("Failed to read schema.sql: %v", err))
	}

	_, err = d.Exec(string(schemaSQL))
	if err != nil {
		panic(fmt.Sprintf("Failed to execute schema.sql: %v", err))
	}

	queries := database.New(d)
	srv := server.New(queries, cfg.EventEmitterBufferLimit, log)
	grpcHandlers := handlers.NewGRPCHandlers(srv)
	httpHandlers := handlers.NewHTTPHandlers(srv)

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
		if err != nil {
			panic(fmt.Sprintf("failed to listen: %v", err))
		}
		s := grpc.NewServer()
		pb.RegisterEventsDBServer(s, grpcHandlers)
		log.Info("gRPC server listening", "address", lis.Addr().String())
		if err := s.Serve(lis); err != nil {
			panic(fmt.Sprintf("failed to serve: %v", err))
		}
	}()

	// Start REST server
	http.HandleFunc("/events", httpHandlers.CreateEventHandler)
	http.HandleFunc("/events/get", httpHandlers.GetEventByIDHandler)
	http.HandleFunc("/events/stream", httpHandlers.StreamEventsFromSubjectHandler)

	log.Info("REST server listening", "address", fmt.Sprintf(":%d", cfg.RESTPort))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.RESTPort), nil); err != nil {
		panic(fmt.Sprintf("failed to serve: %v", err))
	}
}
