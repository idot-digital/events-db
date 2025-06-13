# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Events DB is a high-performance event storage and streaming service built in Go that provides both gRPC and HTTP interfaces for event management. It follows the CloudEvents specification and supports real-time event streaming to multiple clients.

## Architecture

The project uses clean architecture with the following key components:

- **Dual Interfaces**: gRPC server (port 50051) and HTTP REST server (port 8080) running concurrently
- **Database Layer**: MySQL with SQLC for type-safe database access
- **Client Libraries**: TypeScript clients for both gRPC and HTTP interfaces
- **Real-time Streaming**: Server-Sent Events (HTTP) and gRPC streaming with in-memory event distribution
- **Event Management**: Single events table with listeners that broadcast to connected clients

## Essential Commands

### Build Commands
```bash
make build          # Complete build (proto, sqlc, clients, server)
make build-db       # Build main server binary only
make proto          # Generate gRPC code from protobuf
make sqlc           # Generate database access code from queries
make clean          # Clean build artifacts
```

### Client Building
```bash
make build-ts-grpc-client   # Build TypeScript gRPC client
make build-ts-http-client   # Build TypeScript HTTP client
```

### Development
```bash
make run-db                                      # Run server directly with go run
./bin/eventsdb --grpc-port=50051 --rest-port=8080  # Run built binary
```

## Key Architecture Patterns

### Event Distribution
- Events are stored in MySQL and simultaneously broadcast to in-memory listeners
- Each connected client has a buffered channel for event delivery
- Client management handles connection lifecycle and cleanup

### Dual Protocol Support
- Same business logic serves both gRPC and HTTP interfaces
- HTTP uses Server-Sent Events for streaming
- gRPC uses bidirectional streaming

### Database Integration
- SQLC generates type-safe Go code from SQL queries in `database/queries/`
- Schema is defined in `schema.sql` and auto-created on startup
- Single events table with indexes on subject and timestamp

## Important File Locations

- **Proto Definition**: `eventsdb.proto` (generates `grpc/` directory)
- **Database Queries**: `database/queries/events.sql` (generates `database/` package)
- **Server Main**: `cmd/server/main.go`
- **Core Logic**: `internal/server/server.go`
- **OpenAPI Spec**: `docs/openapi.yaml`

## Configuration

The server accepts both environment variables and command-line flags:
- MySQL connection via `MYSQL_*` env vars
- Optional authentication via `AUTH_TOKEN`
- TLS configuration via `TLS_CERT_FILE`/`TLS_KEY_FILE`
- Client limits via `--max-total-clients` and `--client-buffer-size`

## Testing

No test files currently exist in the codebase. Use standard Go testing commands:
```bash
go test ./...           # Run all tests
go test -v ./...        # Run tests with verbose output
go test ./internal/...  # Run tests for specific package
```