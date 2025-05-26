import {
  Event,
  CreateEventRequest,
  CreateEventResponse,
  EventsDBClientConfig,
} from "./types";

import { EventSource } from "eventsource";

export class EventsDBClient {
  private address: string;
  private headers: HeadersInit;

  constructor(config: EventsDBClientConfig) {
    this.address = config.address;
    this.headers = {
      "Content-Type": "application/json",
      ...(config.token ? { Authorization: `Bearer ${config.token}` } : {}),
    };
  }

  /**
   * Create a new event
   * @returns The ID of the created event
   */
  async createEvent(request: CreateEventRequest): Promise<number> {
    const response = await fetch(`${this.address}/events`, {
      method: "POST",
      headers: this.headers,
      body: JSON.stringify({
        source: request.source,
        type: request.type,
        subject: request.subject,
        data: request.data.toString(),
      }),
    });

    if (!response.ok) {
      throw new Error(
        `HTTP error! status: ${response.status} ${await response.text()}`
      );
    }

    return ((await response.json()) as CreateEventResponse).id;
  }

  /**
   * Get an event by ID
   */
  async getEventByID(id: number): Promise<Event> {
    const response = await fetch(`${this.address}/events/get?id=${id}`, {
      method: "GET",
      headers: this.headers,
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response.json();
  }

  /**
   * Stream events for a subject
   * @param subject The subject to stream events for
   * @param onEvent Callback function that will be called for each event
   * @param onError Callback function that will be called if an error occurs
   * @returns A function to stop the stream
   */
  streamEventsFromSubject(
    subject: string,
    onEvent: (event: Event) => void,
    onError: (error: Error) => void
  ): () => void {
    const url = `${this.address}/events/stream?subject=${encodeURIComponent(
      subject
    )}`;
    const eventSource = new EventSource(url);

    eventSource.onmessage = (message) => {
      try {
        const eventData = JSON.parse(message.data) as Event;
        onEvent(eventData);
      } catch (error) {
        onError(
          error instanceof Error
            ? error
            : new Error("Failed to parse event data")
        );
      }
    };

    eventSource.onerror = (error) => {
      onError(error instanceof Error ? error : new Error("Stream error"));
      eventSource.close();
    };

    return () => eventSource.close();
  }
}
