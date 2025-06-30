import {
  EventsDBClientImpl,
  GetEventByIDRequest,
  EventsFromSubjectRequest,
  Event,
  EventsFromSubjectReply,
} from "../gen/eventsdb";
import * as grpc from "@grpc/grpc-js";
import * as fs from "fs";
import { Observable } from "rxjs";

export interface EventsDBClientConfig {
  address: string;
  /** Source of the events. */
  source: string;
  token?: string;
  tlsCertFile?: string;
  tlsKeyFile?: string;
}

export interface CreateEventRequest {
  type: string;
  subject: string;
  data: Uint8Array;
}

export class EventsDBClient {
  private client: EventsDBClientImpl;
  private source: string;

  constructor(config: EventsDBClientConfig) {
    this.source = config.source;
    // Create credentials based on TLS configuration
    let credentials: grpc.ChannelCredentials;
    if (config.tlsCertFile && config.tlsKeyFile) {
      try {
        const rootCert = fs.readFileSync(config.tlsCertFile);
        credentials = grpc.credentials.createSsl(rootCert);
      } catch (error) {
        console.error("Error reading TLS certificate file:", error);
        throw new Error(`Failed to read TLS certificate: ${error}`);
      }
    } else {
      credentials = grpc.credentials.createInsecure();
    }

    // Create an RPC implementation that uses the gRPC client
    const rpc = {
      request: (
        service: string,
        method: string,
        data: Uint8Array
      ): Promise<Uint8Array> => {
        return new Promise((resolve, reject) => {
          const client = new grpc.Client(config.address, credentials);

          // Add authentication metadata if token is provided
          const metadata = new grpc.Metadata();
          if (config.token) {
            metadata.add("authorization", config.token);
          }

          client.makeUnaryRequest(
            `/${service}/${method}`,
            (arg: any) => Buffer.from(arg),
            (buffer: Buffer) => buffer,
            data,
            metadata,
            (err: any, response: any) => {
              if (err) {
                reject(err);
              } else {
                resolve(response);
              }
              client.close();
            }
          );
        });
      },
      clientStreamingRequest: (
        service: string,
        method: string,
        data: any
      ): Promise<Uint8Array> => {
        throw new Error("Client streaming not implemented");
      },
      serverStreamingRequest: (
        service: string,
        method: string,
        data: Uint8Array
      ): Observable<Uint8Array> => {
        const client = new grpc.Client(config.address, credentials);

        // Add authentication metadata if token is provided
        const metadata = new grpc.Metadata();
        if (config.token) {
          metadata.add("authorization", config.token);
        }

        const stream = client.makeServerStreamRequest(
          `/${service}/${method}`,
          (arg: any) => Buffer.from(arg),
          (buffer: Buffer) => buffer,
          data,
          metadata
        );

        return new Observable<Uint8Array>((subscriber) => {
          stream.on("data", (data: Buffer) =>
            subscriber.next(new Uint8Array(data))
          );
          stream.on("error", (err: any) => subscriber.error(err));
          stream.on("end", () => subscriber.complete());
          return () => stream.cancel();
        });
      },
      bidirectionalStreamingRequest: (
        service: string,
        method: string,
        data: any
      ): any => {
        throw new Error("Bidirectional streaming not implemented");
      },
    };

    this.client = new EventsDBClientImpl(rpc, { service: "grpc.EventsDB" });
  }

  public async createEvent(
    request: CreateEventRequest
  ): Promise<number | null> {
    try {
      const response = await this.client.CreateEvent({
        ...request,
        source: this.source,
      });
      return response.id;
    } catch (error) {
      console.error("Error creating event:", error);
      return null;
    }
  }

  public async getEventByID(id: number): Promise<Event | null> {
    try {
      const request: GetEventByIDRequest = { id };
      return await this.client.GetEventByID(request);
    } catch (error) {
      return null;
    }
  }

  public streamEventsFromSubject(
    subject: string,
    options: {
      fromId?: number;
      type?: string;
      recursive?: boolean;
    } = {}
  ): Observable<EventsFromSubjectReply> {
    try {
      const request: EventsFromSubjectRequest = {
        subject,
        fromId: options.fromId,
        type: options.type,
        recursive: options.recursive,
      };
      return this.client.StreamEventsFromSubject(request);
    } catch (error) {
      console.error("Error streaming events from subject:", error);
      return new Observable<EventsFromSubjectReply>((subscriber) => {
        subscriber.error(error);
      });
    }
  }

  public async getHistoricEventsFromSubject(
    subject: string,
    options: {
      fromId?: number;
      type?: string;
      recursive?: boolean;
    } = {}
  ): Promise<EventsFromSubjectReply | null> {
    try {
      const request: EventsFromSubjectRequest = {
        subject,
        fromId: options.fromId ?? undefined,
        type: options.type,
        recursive: options.recursive,
      };
      return await this.client.GetHistoricEventsFromSubject(request);
    } catch (error) {
      console.error("Error getting historic events from subject:", error);
      return null;
    }
  }

  public async deleteFromSubject(
    subject: string,
    options: {
      fromId?: number;
      type?: string;
      recursive?: boolean;
    } = {}
  ): Promise<boolean> {
    try {
      const request: EventsFromSubjectRequest = {
        subject,
        fromId: options.fromId ?? undefined,
        type: options.type,
        recursive: options.recursive,
      };
      const response = await this.client.DeleteFromSubject(request);
      return response.ok;
    } catch (error) {
      console.error("Error deleting from subject:", error);
      return false;
    }
  }
}
