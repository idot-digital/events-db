import {
  EventsDBClientImpl,
  GetEventByIDRequest,
  EventsFromSubjectRequest,
  Event,
  EventsFromSubjectReply,
  OkReply,
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
      const rootCert = fs.readFileSync(config.tlsCertFile);
      credentials = grpc.credentials.createSsl(rootCert);
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

  async createEvent(request: CreateEventRequest): Promise<number> {
    const response = await this.client.CreateEvent({
      ...request,
      source: this.source,
    });
    return response.id;
  }

  async getEventByID(id: number): Promise<Event> {
    const request: GetEventByIDRequest = { id };
    return await this.client.GetEventByID(request);
  }

  streamEventsFromSubject(
    subject: string,
    options: {
      fromId?: number;
      type?: string;
      recursive?: boolean;
    } = {}
  ): Observable<EventsFromSubjectReply> {
    const request: EventsFromSubjectRequest = {
      subject,
      fromId: options.fromId,
      type: options.type,
      recursive: options.recursive,
    };
    return this.client.StreamEventsFromSubject(request);
  }

  getHistoricEventsFromSubject(
    subject: string,
    options: {
      fromId?: number;
      type?: string;
      recursive?: boolean;
    } = {}
  ): Promise<EventsFromSubjectReply> {
    const request: EventsFromSubjectRequest = {
      subject,
      fromId: options.fromId ?? undefined,
      type: options.type,
      recursive: options.recursive,
    };
    return this.client.GetHistoricEventsFromSubject(request);
  }

  async deleteFromSubject(
    subject: string,
    options: {
      fromId?: number;
      type?: string;
      recursive?: boolean;
    } = {}
  ): Promise<boolean> {
    const request: EventsFromSubjectRequest = {
      subject,
      fromId: options.fromId ?? undefined,
      type: options.type,
      recursive: options.recursive,
    };
    const response = await this.client.DeleteFromSubject(request);
    return response.ok;
  }
}
