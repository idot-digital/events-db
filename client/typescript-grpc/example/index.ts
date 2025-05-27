import { EventsDBClient } from "@idot-digital/eventsdb-grpc-client";

const client = new EventsDBClient({
  address: "localhost:50051",
});

const eventID = await client.createEvent({
  source: "source",
  type: "type",
  subject: "subject",
  data: Buffer.from("😀", "utf-8"),
});

console.log(`Event created with ID: ${eventID}`);

const event = await client.getEventByID(eventID);

console.log(
  `Got event with ID: ${event.id}: ${JSON.stringify({
    id: event.id,
    source: event.source,
    type: event.type,
    subject: event.subject,
    time: event.time,
    data: Buffer.from(event.data).toString(),
  })}`
);

const stream = client.streamEventsFromSubject("subject");

const subscription = stream.subscribe((event) => {
  event.events.forEach((e) => {
    console.log(
      `Received event: ${e.id}: ${JSON.stringify({
        id: e.id,
        source: e.source,
        type: e.type,
        subject: e.subject,
        time: e.time,
        data: Buffer.from(e.data).toString(),
      })}`
    );
  });
});

setTimeout(() => {
  subscription.unsubscribe();
  console.log("Subscription cancelled");
}, 1000);
