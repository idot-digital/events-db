.PHONY: build clean proto sqlc

# Build the eventsdb binary
build:
	go build -o bin/eventsdb eventsdb.go

# Clean build artifacts
clean:
	rm -rf bin/

# Create bin directory if it doesn't exist
bin:
	mkdir -p bin

# Generate gRPC code from proto file
proto:
	protoc --go_out=./grpc --go_opt=paths=source_relative --go-grpc_out=./grpc --go-grpc_opt=paths=source_relative eventsdb.proto

# Generate SQL code using sqlc
sqlc:
	mkdir database && sqlc generate