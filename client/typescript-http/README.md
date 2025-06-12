# Events DB TypeScript HTTP Client

A TypeScript client for the Events DB API that provides a simple and type-safe interface to interact with the events database via HTTP. This client library makes it easy to create, retrieve, and stream events while providing full TypeScript support and comprehensive error handling.

## Installation

```bash
npm install @idot-digital/eventsdb-http-client
```

## Quick Start

```typescript
import { EventsDBClient } from "@idot-digital/eventsdb-http-client";

// Create a new client instance
const client = new EventsDBClient({
  address: "http://localhost:8080",
  source: "my-app",
  token: "your-auth-token", // Only needed if authentication is enabled
});

// Create a new event
const eventId = await client.createEvent({
  type: "user.login",
  subject: "user-123",
  data: Buffer.from('{"service": "my-app", "user_id": "user-123", "ip": "127.0.0.1"}'),
});

// Get an event by ID
const retrievedEvent = await client.getEventByID(eventId);

// Stream events for a subject
const stopStream = client.streamEventsFromSubject(
  "user-123",
  (event) => {
    console.log("Received event:", event);
  },
  (error) => {
    console.error("Stream error:", error);
  }
);

// Stop streaming when done
stopStream();
```

## API Reference

### `EventsDBClient`

#### Constructor

```typescript
constructor(config: EventsDBClientConfig)
```

Creates a new client instance. The config object accepts the following properties:

- `address`: The server address including protocol (e.g., "http://localhost:8080") (required)
- `source`: The source identifier for events created by this client (required)
- `token`: Authentication token (optional)

#### Methods

##### `createEvent(request: CreateEventRequest): Promise<number>`

Creates a new event and returns its ID. The request object should contain:
- `type`: Event type (required)
- `subject`: Event subject (required)
- `data`: Event data as Buffer (required)

##### `getEventByID(id: number): Promise<Event>`

Retrieves an event by its ID. Returns an Event object with:
- `id`: Event ID
- `source`: Event source
- `type`: Event type
- `subject`: Event subject
- `time`: Event timestamp
- `data`: Event data as Uint8Array

##### `streamEventsFromSubject(subject: string, onEvent: (event: Event) => void, onError: (error: Error) => void): () => void`

Streams events for a given subject using Server-Sent Events. Returns a function to stop the stream.

- `subject`: The subject to stream events for
- `onEvent`: Callback function called for each received event
- `onError`: Callback function called when an error occurs

## Authentication

The client supports Bearer token authentication. Include the `token` property in the client configuration, and it will be automatically added to all requests as an `Authorization: Bearer <token>` header.

## Error Handling

The client throws errors for HTTP failures and provides detailed error messages including the HTTP status code and response text when available. When streaming, errors are passed to the `onError` callback function.
