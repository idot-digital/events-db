export interface Event {
  id: number;
  source: string;
  type: string;
  subject: string;
  time: string;
  data: Uint8Array<ArrayBufferLike>;
}

export interface CreateEventRequest {
  type: string;
  subject: string;
  data: Buffer;
}

export interface CreateEventResponse {
  id: number;
}

export interface EventsDBClientConfig {
  address: string;
  /** Source of the events. */
  source: string;
  token?: string;
}

export interface EventsFromSubjectReply {
  events: Event[];
  has_more: boolean;
}

export interface GetHistoricEventsFromSubjectInternalReply {
  events: {
    id: number;
    source: string;
    type: string;
    subject: string;
    time: string;
    data: string;
  }[];
  has_more: boolean;
}
