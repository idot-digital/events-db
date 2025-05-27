.PHONY: build clean proto sqlc build-db build-ts-grpc-client build-ts-http-client

build: proto sqlc build-ts-grpc-client build-ts-http-client build-db

clean:
	rm -rf bin/

bin:
	mkdir -p bin

proto:
	mkdir -p grpc && protoc --go_out=./grpc --go_opt=paths=source_relative --go-grpc_out=./grpc --experimental_allow_proto3_optional --go-grpc_opt=paths=source_relative eventsdb.proto

sqlc:
	mkdir -p database && sqlc generate

build-db:
	go build -o bin/eventsdb cmd/server/main.go

ts-client-install:
	cd client/typescript-grpc && npm install
	cd client/typescript-http && npm install

build-ts-grpc-client: ts-client-install
	cd client/typescript-grpc && rm -rf gen && rm -rf dist && npm run build

build-ts-http-client: ts-client-install
	cd client/typescript-http && rm -rf dist && npm run build