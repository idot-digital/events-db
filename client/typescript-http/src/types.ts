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
  data: string;
}

export interface CreateEventResponse {
  id: number;
}

export interface EventsDBClientConfig {
  baseURL: string;
  token?: string;
}
