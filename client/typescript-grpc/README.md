# EventsDB gRPC TypeScript Client

A Node.js TypeScript client for EventsDB. It uses the gRPC protocol to communicate with the EventsDB server.

## Installation

```bash
npm install
```

## Building

First, generate the TypeScript types from the proto file:

```bash
npm run generate
```

Then build the project:

```bash
npm run build
```

## Usage

```typescript
import { EventsDBClient } from "./dist";

// Create a client instance without authentication or TLS
const client = new EventsDBClient({
  address: "localhost:50051",
});

// Create a client with authentication
const clientWithAuth = new EventsDBClient({
  address: "localhost:50051",
  token: "your-auth-token",
});

// Create a client with both authentication and TLS
const clientWithTLS = new EventsDBClient({
  address: "localhost:50051",
  token: "your-auth-token",
  tlsCertFile: "/path/to/cert.pem",
  tlsKeyFile: "/path/to/key.pem",
});

// Create an event
const eventId = await client.createEvent(
  "test-source",
  "test-type",
  "test-subject",
  Buffer.from("test-data")
);
console.log("Created event with ID:", eventId);

// Get an event
const event = await client.getEventByID(1);
console.log("Retrieved event:", event);

// Stream events
const stream = client.streamEventsFromSubject("test-subject", {
  recursive: true,
});
stream.subscribe({
  next: (response) => {
    console.log("Received events:", response.events);
  },
  error: (error) => {
    console.error("Stream error:", error);
  },
  complete: () => {
    console.log("Stream completed");
  },
});
```

## API Reference

### `EventsDBClient`

#### Constructor

```typescript
constructor(config: EventsDBClientConfig)
```

Creates a new client instance. The config object accepts the following properties:

- `address`: The server address in the format `host:port` (required)
- `token`: Authentication token (optional)
- `tlsCertFile`: Path to TLS certificate file (optional)
- `tlsKeyFile`: Path to TLS key file (optional)

#### Methods

##### `createEvent(source: string, type: string, subject: string, data: Buffer): Promise<number>`

Creates a new event and returns its ID.

##### `getEventByID(id: number): Promise<Event>`

Retrieves an event by its ID.

##### `streamEventsFromSubject(subject: string, options?: { type?: string; fromId?: number; recursive?: boolean }): Observable<StreamEventsFromSubjectReply>`

Streams events for a given subject. The options parameter is optional and can include:

- `type`: Filter events by type
- `fromId`: Start streaming from a specific event ID
- `recursive`: Whether to include events from child subjects

## Authentication

The client supports token-based authentication. To use authentication:

1. Set the `token` property in the client configuration
2. The token will be automatically added to all requests as an `authorization` header

## TLS Support

The client supports TLS connections. To use TLS:

1. Provide the paths to your TLS certificate and key files in the client configuration
2. The client will automatically establish a secure connection using these credentials

Note: If TLS files are not provided, the client will use an insecure connection.
