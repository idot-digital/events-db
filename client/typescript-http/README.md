# Events DB TypeScript HTTP Client

A TypeScript client for the Events DB API that provides a simple and type-safe interface to interact with the events database via HTTP. This client library makes it easy to create, retrieve, and stream events while providing full TypeScript support and comprehensive error handling.

## Installation

```bash
npm install @idot-digital/events-db-http-client
```

## Quick Start

```typescript
import { EventsDBClient } from "@idot-digital/events-db-http-client";

// Create a new client instance
const client = new EventsDBClient({
  baseURL: "http://localhost:8080",
  token: "your-auth-token", // Only needed, if authentication is enabled
});

// Create a new event
const event = await client.createEvent({
  source: "my-app",
  type: "user.login",
  subject: "user-123",
  data: '{"service": "my-app", "user_id": "user-123", "ip": "127.0.0.1"}',
});

// Get an event by ID
const retrievedEvent = await client.getEvent(event.id);

// Stream events for a subject
const stopStream = client.streamEvents(
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
