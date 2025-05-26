export interface Event {
  id: number;
  source: string;
  type: string;
  subject: string;
  time: string;
  data: string;
}

export interface CreateEventRequest {
  source: string;
  type: string;
  subject: string;
  data: Buffer;
}

export interface CreateEventResponse {
  id: number;
}

export interface EventsDBClientConfig {
  address: string;
  token?: string;
}
