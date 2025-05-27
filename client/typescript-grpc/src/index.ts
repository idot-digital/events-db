import {
  EventsDBClientImpl,
  CreateEventRequest,
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
  token?: string;
  tlsCertFile?: string;
  tlsKeyFile?: string;
}

export class EventsDBClient {
  private client: EventsDBClientImpl;

  constructor(config: EventsDBClientConfig) {
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
    const response = await this.client.CreateEvent(request);
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
      // type?: string;
      // recursive?: boolean;
    } = {}
  ): Observable<EventsFromSubjectReply> {
    const request: EventsFromSubjectRequest = {
      subject,
      fromId: options.fromId,
      // type: options.type,
      // recursive: options.recursive,
    };
    return this.client.StreamEventsFromSubject(request);
  }

  getHistoricEventsFromSubject(
    subject: string,
    options: {
      fromId?: number;
      // type?: string;
      // recursive?: boolean;
    } = {}
  ): Promise<EventsFromSubjectReply> {
    const request: EventsFromSubjectRequest = {
      subject,
      fromId: options.fromId ?? undefined,
    };
    return this.client.GetHistoricEventsFromSubject(request);
  }
}
