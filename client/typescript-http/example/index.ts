import { EventsDBClient } from "@idot-digital/eventsdb-http-client";

const client = new EventsDBClient({
  address: "http://localhost:8080",
});

const eventID = await client.createEvent({
  source: "source",
  type: "type",
  subject: "subject",
  data: Buffer.from("😁"),
});

console.log(`Event created with ID: ${eventID}`);

const event = await client.getEventByID(eventID);

console.log(`Got event with ID: ${event.id}: ${JSON.stringify(event)}`);

const unsubscribe = client.streamEventsFromSubject(
  "subject",
  (event) => {
    console.log(
      `Received event: ${event.id}: ${JSON.stringify({
        id: event.id,
        source: event.source,
        type: event.type,
        subject: event.subject,
        time: event.time,
        data: Buffer.from(event.data).toString(),
      })}`
    );
  },
  (error) => console.error(error)
);

setTimeout(() => {
  unsubscribe();
  console.log("Subscription cancelled");
}, 1000);
