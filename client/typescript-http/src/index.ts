import {
  Event,
  CreateEventRequest,
  CreateEventResponse,
  EventsDBClientConfig,
  EventsFromSubjectReply,
  GetHistoricEventsFromSubjectInternalReply,
} from "./types";

import { EventSource } from "eventsource";

export class EventsDBClient {
  private address: string;
  private headers: HeadersInit;
  private source: string;

  constructor(config: EventsDBClientConfig) {
    this.address = config.address;
    this.headers = {
      "Content-Type": "application/json",
      ...(config.token ? { Authorization: `Bearer ${config.token}` } : {}),
    };
    this.source = config.source;
  }

  /**
   * Create a new event
   * @returns The ID of the created event or null if error
   */
  async createEvent(request: CreateEventRequest): Promise<number | null> {
    try {
      const response = await fetch(`${this.address}/events`, {
        method: "POST",
        headers: this.headers,
        body: JSON.stringify({
          source: this.source,
          type: request.type,
          subject: request.subject,
          data: request.data.toString("base64"),
        }),
      });

      if (!response.ok) {
        console.error(`HTTP error! status: ${response.status} ${await response.text()}`);
        return null;
      }

      return ((await response.json()) as CreateEventResponse).id;
    } catch (error) {
      console.error("Error creating event:", error);
      return null;
    }
  }

  /**
   * Get an event by ID
   */
  async getEventByID(id: number): Promise<Event | null> {
    try {
      const response = await fetch(`${this.address}/events/get?id=${id}`, {
        method: "GET",
        headers: this.headers,
      });

      if (!response.ok) {
        console.error(`HTTP error! status: ${response.status}`);
        return null;
      }

      const data = await response.json();

      return {
        id: data.id,
        source: data.source,
        type: data.type,
        subject: data.subject,
        time: data.time,
        data: Buffer.from(data.data, "base64"),
      };
    } catch (error) {
      console.error("Error getting event by ID:", error);
      return null;
    }
  }

  /**
   * Stream events for a subject
   * @param subject The subject to stream events for
   * @param onEvent Callback function that will be called for each event
   * @param onError Callback function that will be called if an error occurs
   * @returns A function to stop the stream or null if initialization fails
   */
  streamEventsFromSubject(
    subject: string,
    options: {
      fromId?: number;
      type?: string;
      recursive?: boolean;
    } = {},
    onEvent: (event: Event) => void,
    onError: (error: Error) => void
  ): (() => void) | null {
    try {
      const url = new URL(`${this.address}/events/stream`);
      url.searchParams.set("subject", subject);
      if (options.fromId) {
        url.searchParams.set("from_id", options.fromId.toString());
      }
      if (options.type) {
        url.searchParams.set("type", options.type);
      }
      if (options.recursive) {
        url.searchParams.set("recursive", "true");
      }
      const eventSource = new EventSource(url);

      eventSource.onmessage = (message) => {
        try {
          const rawEventData = JSON.parse(message.data);
          const eventData = {
            id: rawEventData.id,
            source: rawEventData.source,
            type: rawEventData.type,
            subject: rawEventData.subject,
            time: rawEventData.time,
            data: Buffer.from(rawEventData.data, "base64"),
          };
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
    } catch (error) {
      console.error("Error initializing event stream:", error);
      onError(error instanceof Error ? error : new Error("Stream initialization failed"));
      return null;
    }
  }

  async getHistoricEventsFromSubject(
    subject: string,
    options: {
      fromId?: number;
      type?: string;
      recursive?: boolean;
    } = {}
  ): Promise<EventsFromSubjectReply | null> {
    try {
      const url = new URL(`${this.address}/events/historic`);
      url.searchParams.set("subject", subject);
      if (options.fromId) {
        url.searchParams.set("from_id", options.fromId.toString());
      }
      if (options.type) {
        url.searchParams.set("type", options.type);
      }
      if (options.recursive) {
        url.searchParams.set("recursive", "true");
      }
      
      const fetchResponse = await fetch(url, {
        method: "GET",
        headers: this.headers,
      });
      
      if (!fetchResponse.ok) {
        console.error(`HTTP error! status: ${fetchResponse.status}`);
        return null;
      }
      
      const response = (await fetchResponse.json()) as GetHistoricEventsFromSubjectInternalReply;

      const result: EventsFromSubjectReply = {
        events: response.events.map((e) => ({
          id: e.id,
          source: e.source,
          type: e.type,
          subject: e.subject,
          time: e.time,
          data: Buffer.from(e.data, "base64"),
        })),
        has_more: response.has_more,
      };
      return result;
    } catch (error) {
      console.error("Error getting historic events from subject:", error);
      return null;
    }
  }

  /**
   * Delete events from a subject
   * @param subject The subject to delete events from
   * @param options Optional parameters including fromId
   * @returns Promise<boolean> indicating success
   */
  async deleteFromSubject(
    subject: string,
    options: {
      fromId?: number;
      type?: string;
      recursive?: boolean;
    } = {}
  ): Promise<boolean> {
    try {
      const url = new URL(`${this.address}/subjects/delete`);
      url.searchParams.set("subject", subject);
      if (options.fromId) {
        url.searchParams.set("from_id", options.fromId.toString());
      }
      if (options.type) {
        url.searchParams.set("type", options.type);
      }
      if (options.recursive) {
        url.searchParams.set("recursive", "true");
      }

      const response = await fetch(url, {
        method: "POST",
        headers: this.headers,
      });

      if (!response.ok) {
        console.error(`HTTP error! status: ${response.status} ${await response.text()}`);
        return false;
      }

      const result = await response.json();
      return result.status === "ok";
    } catch (error) {
      console.error("Error deleting from subject:", error);
      return false;
    }
  }
}
