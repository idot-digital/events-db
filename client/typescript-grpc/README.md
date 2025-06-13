# EventsDB gRPC TypeScript Client

A Node.js TypeScript client for EventsDB. It uses the gRPC protocol to communicate with the EventsDB server.

## Installation

```bash
npm install @idot-digital/eventsdb-grpc-client
```

## Usage

```typescript
import { EventsDBClient } from "@idot-digital/eventsdb-grpc-client";

// Create a client instance without authentication or TLS
const client = new EventsDBClient({
  address: "localhost:50051",
  source: "my-app",
});

// Create a client with authentication
const clientWithAuth = new EventsDBClient({
  address: "localhost:50051",
  source: "my-app",
  token: "your-auth-token",
});

// Create a client with both authentication and TLS
const clientWithTLS = new EventsDBClient({
  address: "localhost:50051",
  source: "my-app",
  token: "your-auth-token",
  tlsCertFile: "/path/to/cert.pem",
  tlsKeyFile: "/path/to/key.pem",
});

// Create an event
const eventId = await client.createEvent({
  type: "user.login",
  subject: "user-123",
  data: Buffer.from(
    '{"service": "my-app", "user_id": "user-123", "ip": "127.0.0.1"}'
  ),
});
console.log("Created event with ID:", eventId);

// Get an event
const event = await client.getEventByID(eventId);
console.log("Retrieved event:", event);

// Delete events from a subject
const success = await client.deleteFromSubject("user-123", {
  fromId: 100, // Delete events with ID >= 100
});
console.log("Delete operation success:", success);

// Delete events recursively (from subjects starting with "user-")
const recursiveSuccess = await client.deleteFromSubject("user-", {
  fromId: 0,
  recursive: true, // Deletes from "user-123", "user-456", etc.
});

// Delete events of specific type only
const typeSuccess = await client.deleteFromSubject("user-123", {
  type: "user.login", // Only delete login events
});

// Stream events
const stream = client.streamEventsFromSubject("user-123", {
  fromId: 1,
});
stream.subscribe({
  next: (response) => {
    response.events.forEach((event) => {
      console.log("Received event:", event);
    });
  },
  error: (error) => {
    console.error("Stream error:", error);
  },
  complete: () => {
    console.log("Stream completed");
  },
});

// Stream events with type filtering
const typeFilteredStream = client.streamEventsFromSubject("user-123", {
  fromId: 1,
  type: "user.login", // Only stream login events
});

// Stream events recursively (from multiple subjects)
const recursiveStream = client.streamEventsFromSubject("user-", {
  fromId: 1,
  recursive: true, // Streams from "user-123", "user-456", etc.
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
- `source`: The source identifier for events created by this client (required)
- `token`: Authentication token (optional)
- `tlsCertFile`: Path to TLS certificate file (optional)
- `tlsKeyFile`: Path to TLS key file (optional)

#### Methods

##### `createEvent(request: CreateEventRequest): Promise<number>`

Creates a new event and returns its ID. The request object should contain:

- `type`: Event type (required)
- `subject`: Event subject (required)
- `data`: Event data as Uint8Array (required)

##### `getEventByID(id: number): Promise<Event>`

Retrieves an event by its ID.

##### `streamEventsFromSubject(subject: string, options?: { fromId?: number, type?: string, recursive?: boolean }): Observable<StreamEventsFromSubjectReply>`

Streams events for a given subject. The options parameter is optional and can include:

- `fromId`: Start streaming from a specific event ID
- `type`: Filter events by specific type
- `recursive`: When `true`, matches subjects that start with the given subject (e.g., "test" matches "test", "test/1", "test/sub/path")

##### `getHistoricEventsFromSubject(subject: string, options?: { fromId?: number, type?: string, recursive?: boolean }): Promise<EventsFromSubjectReply>`

Retrieves historic events for a given subject. The options parameter is optional and can include:

- `fromId`: Start from a specific event ID
- `type`: Filter events by specific type
- `recursive`: When `true`, matches subjects that start with the given subject

##### `deleteFromSubject(subject: string, options?: { fromId?: number, type?: string, recursive?: boolean }): Promise<boolean>`

Deletes events from a given subject. The options parameter is optional and can include:

- `fromId`: Delete events with ID greater than or equal to this value (defaults to 0 if not provided)
- `type`: Only delete events of the specified type
- `recursive`: When `true`, deletes events from subjects that start with the given subject

Returns `true` on successful deletion.

## Authentication

The client supports token-based authentication. To use authentication:

1. Set the `token` property in the client configuration
2. The token will be automatically added to all requests as an `authorization` header

## TLS Support

The client supports TLS connections. To use TLS:

1. Provide the paths to your TLS certificate and key files in the client configuration
2. The client will automatically establish a secure connection using these credentials

Note: If TLS files are not provided, the client will use an insecure connection.
