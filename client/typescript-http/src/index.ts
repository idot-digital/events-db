import {
  Event,
  CreateEventRequest,
  CreateEventResponse,
  EventsDBClientConfig,
} from "./types";

export class EventsDBClient {
  private baseURL: string;
  private headers: HeadersInit;

  constructor(config: EventsDBClientConfig) {
    this.baseURL = config.baseURL;
    this.headers = {
      "Content-Type": "application/json",
      ...(config.token ? { Authorization: `Bearer ${config.token}` } : {}),
    };
  }

  /**
   * Create a new event
   */
  async createEvent(request: CreateEventRequest): Promise<CreateEventResponse> {
    const response = await fetch(`${this.baseURL}/events`, {
      method: "POST",
      headers: this.headers,
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response.json();
  }

  /**
   * Get an event by ID
   */
  async getEvent(id: number): Promise<Event> {
    const response = await fetch(`${this.baseURL}/events/get?id=${id}`, {
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
  streamEvents(
    subject: string,
    onEvent: (event: Event) => void,
    onError: (error: Error) => void
  ): () => void {
    const url = `${this.baseURL}/events/stream?subject=${encodeURIComponent(
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
