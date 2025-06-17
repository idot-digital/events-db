FROM --platform=$BUILDPLATFORM alpine:latest as builder

ARG TARGETPLATFORM
ARG BINARY_AMD64
ARG BINARY_ARM64

WORKDIR /app

# Copy the binaries into the build context
COPY $BINARY_AMD64 /app/eventsdb-amd64
COPY $BINARY_ARM64 /app/eventsdb-arm64

# Select the appropriate binary based on the target platform
RUN if [ "$TARGETPLATFORM" = "linux/amd64" ]; then \
        cp /app/eventsdb-amd64 /app/eventsdb; \
    elif [ "$TARGETPLATFORM" = "linux/arm64" ]; then \
        cp /app/eventsdb-arm64 /app/eventsdb; \
    else \
        echo "Unsupported platform: $TARGETPLATFORM" && exit 1; \
    fi

FROM --platform=$TARGETPLATFORM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl tar sudo

RUN curl -s https://vault-loader.idot-digital.com/install.sh | sh

# Copy the pre-built binary
COPY --from=builder /app/eventsdb /app/eventsdb

# Copy schema and queries
COPY schema.sql /app/schema.sql
COPY query.sql /app/query.sql

# Expose the application port
EXPOSE 8080
EXPOSE 50051

# Run the application
ENTRYPOINT [ "vault-loader", "run", "--ignore-if-fail", "/app/eventsdb" ]