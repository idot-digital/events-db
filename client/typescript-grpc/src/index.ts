import {
  EventsDBClientImpl,
  CreateEventRequest,
  GetEventByIDRequest,
  StreamEventsFromSubjectRequest,
  Event,
} from "../gen/eventsdb";
import * as grpc from "@grpc/grpc-js";
import * as fs from "fs";

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
    const rpc: any = {
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
      ): any => {
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
        return stream;
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

  async createEvent(
    source: string,
    type: string,
    subject: string,
    data: Buffer
  ): Promise<number> {
    const request: CreateEventRequest = {
      source,
      type,
      subject,
      data: new Uint8Array(data),
    };
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
      type?: string;
      fromId?: number;
      recursive?: boolean;
    } = {}
  ): any {
    const request: StreamEventsFromSubjectRequest = {
      subject,
      type: options.type,
      fromId: options.fromId,
      recursive: options.recursive,
    };
    return this.client.StreamEventsFromSubject(request);
  }
}
