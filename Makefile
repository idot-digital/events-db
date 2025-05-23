.PHONY: build clean proto sqlc

# Build the eventsdb binary
build: proto sqlc
	go build -o bin/eventsdb cmd/server/main.go


clean:
	rm -rf bin/


bin:
	mkdir -p bin


proto:
	protoc --go_out=./grpc --go_opt=paths=source_relative --go-grpc_out=./grpc --go-grpc_opt=paths=source_relative eventsdb.proto

sqlc:
	mkdir -p database && sqlc generate