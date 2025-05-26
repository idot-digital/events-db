# Events DB

A high-performance event storage and streaming service that supports both HTTP and gRPC interfaces. The service follows the [CloudEvents](https://cloudevents.io/) specification for event data structure.

## Features

- Event storage and retrieval
- Real-time event streaming using Server-Sent Events (SSE)
- Dual interface support (HTTP REST and gRPC)
- Authentication support
- Prometheus metrics
- TLS support
- MySQL backend
- CloudEvents compliant

## API Documentation

The API is documented using OpenAPI 3.1.0 specification. You can find the complete API documentation in [`docs/openapi.yaml`](docs/openapi.yaml).

### Event Structure

Events follow the CloudEvents specification with the following fields:

| Field   | Type   | Required | Description                  |
| ------- | ------ | -------- | ---------------------------- |
| source  | string | Yes      | Source of the event          |
| type    | string | Yes      | Type of the event            |
| subject | string | Yes      | Subject of the event         |
| data    | bytes  | Yes      | Event data in bytes          |
| time    | string | No       | Time when the event occurred |

## Installation

There are three ways to install and run events-db:

1. Using Docker
2. Using the pre-built binaries
3. Building from source

### Using Docker

The easiest way to run events-db is using Docker. Our example [`docker-compose.yml`](docs/docker-compose.yml) should get you started quickly.

### Using Pre-built Binaries

<!--
The pre-built binaries can be found in the [releases page](https://github.com/idot-digital/events-db/releases). -->

We currently do not have pre-built binaries available. We are working on it!

### Building from Source

```bash
git clone https://github.com/idot-digital/events-db.git
cd events-db
make build
```

## Usage

### HTTP Endpoints

#### Create Event

```http
POST /events
Content-Type: application/json
Authorization: Bearer <token>

{
  "source": "string",
  "type": "string",
  "subject": "string",
  "data": "bytes"
}
```

#### Get Event by ID

```http
GET /events/get?id=<event_id>
Authorization: Bearer <token>
```

#### Stream Events

```http
GET /events/stream?subject=<subject>
Authorization: Bearer <token>
```

### gRPC Interface

The service also provides a gRPC interface with the following methods:

- `CreateEvent`
- `GetEventByID`
- `StreamEventsFromSubject`

The gRPC service definition can be found in `eventsdb.proto`.

## Configuration

The service can be configured using environment variables and command-line flags:

### Environment Variables

- `MYSQL_USER` - MySQL username (default: "root")
- `MYSQL_PASSWORD` - MySQL password (default: "root")
- `MYSQL_DATABASE_NAME` - MySQL database name (default: "root")
- `MYSQL_HOST` - MySQL host (default: "localhost")
- `MYSQL_PORT` - MySQL port (default: "3306")
- `AUTH_TOKEN` - Authentication token (optional)
- `TLS_CERT_FILE` - Path to TLS certificate file (optional)
- `TLS_KEY_FILE` - Path to TLS key file (optional)

### Command-line Flags

- `--grpc-port` - gRPC server port (default: 50051)
- `--rest-port` - REST server port (default: 8080)
- `--client-buffer-size` - Buffer size for client event channels (default: 100)
- `--max-total-clients` - Maximum total number of clients across all subjects (default: 10000)
- `--stream-batch-size` - Number of events to fetch in each stream batch (default: 10)

## Database Schema

The service uses a MySQL database with the following schema:

```sql
CREATE TABLE events (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  source VARCHAR(255) NOT NULL,
  type VARCHAR(255) NOT NULL,
  subject VARCHAR(255) NOT NULL,
  data BLOB NOT NULL,
  time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_subject (subject),
  FULLTEXT INDEX idx_subject_ft (subject)
);
```

## Metrics

The service exposes Prometheus metrics at the `/metrics` endpoint:

- `app_event_operations_total` - Total number of event operations
- `app_event_operation_duration_seconds` - Duration of event operations
- `app_active_event_streams` - Number of currently active event streams

## Security

- Authentication is optional and can be enabled by setting the `AUTH_TOKEN` environment variable
- TLS support can be enabled by providing certificate and key files
- Both HTTP and gRPC interfaces support authentication
- The metrics endpoint is publicly accessible without authentication

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
