FROM idotdigital/builder as cert-generator

RUN openssl req -x509 -nodes -newkey rsa:2048 -keyout server.key -out server.crt -days 365 -subj "/CN=localhost"

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
RUN apk add --no-cache ca-certificates tzdata

# Copy the pre-built binary
COPY --from=builder /app/eventsdb /app/eventsdb

RUN mkdir -p /app/example
COPY --from=cert-generator ./server.key /app/example/server.key
COPY --from=cert-generator ./server.crt /app/example/server.crt
RUN chmod 644 /app/example/server.key /app/example/server.crt

# Copy schema and queries
COPY schema.sql /app/schema.sql
COPY query.sql /app/query.sql

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["/app/eventsdb"] 