package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/idot-digital/events-db/database"
	pb "github.com/idot-digital/events-db/grpc"
	"google.golang.org/grpc"
)

const (
	dbUser         = "post-sqlc"
	dbPassword     = "post-sqlc"
	db             = "post-sqlc"
	dbRootPassword = "db-root-password"
	dbItemLimit    = 10
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedEventsDBServer
	queries *database.Queries
	// Channel to broadcast new events to all active streams
	eventChan chan *pb.Event
}

func (s *server) CreateEvent(ctx context.Context, req *pb.CreateEventRequest) (*pb.CreateEventReply, error) {
	id, err := s.queries.CreateEvent(ctx, database.CreateEventParams{
		Source:  req.Source,
		Type:    req.Type,
		Subject: req.Subject,
		Data:    req.Data,
	})
	if err != nil {
		return nil, err
	}

	// Create the event for broadcasting
	event := &pb.Event{
		Id:      id,
		Source:  req.Source,
		Type:    req.Type,
		Subject: req.Subject,
		Time:    time.Now().Format(time.RFC3339),
		Data:    req.Data,
	}

	// Send event to the channel - this will block until the event is processed
	s.eventChan <- event

	return &pb.CreateEventReply{
		Id: id,
	}, nil
}

func (s *server) GetEventByID(ctx context.Context, req *pb.GetEventByIDRequest) (*pb.Event, error) {
	res, err := s.queries.GetEventByID(ctx, database.GetEventByIDParams{
		ID: req.Id,
	})
	if err != nil {
		return nil, err
	}
	time, err := res.Time.MarshalText()
	if err != nil {
		return nil, err
	}
	return &pb.Event{
		Id:      res.ID,
		Source:  res.Source,
		Type:    res.Type,
		Subject: res.Subject,
		Time:    string(time),
		Data:    res.Data,
	}, nil
}

func (s *server) StreamEventsFromSubject(req *pb.StreamEventsFromSubjectRequest, stream pb.EventsDB_StreamEventsFromSubjectServer) error {
	ctx := stream.Context()

	lastID := int64(0)

	for {
		// Get events for the subject
		events, err := s.queries.GetEventsBySubject(ctx, database.GetEventsBySubjectParams{
			Subject: req.Subject,
			Limit:   dbItemLimit,
			ID:      lastID,
		})
		if err != nil {
			return fmt.Errorf("failed to get events: %v", err)
		}
		// Convert events to protobuf format and send historical events
		pbEvents := make([]*pb.Event, 0, len(events))
		for _, event := range events {
			time, err := event.Time.MarshalText()
			if err != nil {
				return fmt.Errorf("failed to marshal time: %v", err)
			}

			pbEvents = append(pbEvents, &pb.Event{
				Id:      event.ID,
				Source:  event.Source,
				Type:    event.Type,
				Subject: event.Subject,
				Time:    string(time),
				Data:    event.Data,
			})
		}
		// Send historical events
		if len(pbEvents) > 0 {
			reply := &pb.StreamEventsFromSubjectReply{
				Events: pbEvents,
			}
			lastID = pbEvents[len(pbEvents)-1].Id
			if err := stream.Send(reply); err != nil {
				return fmt.Errorf("failed to send historical events: %v", err)
			}
		} else {
			break
		}
	}

	// Listen for new events
	for {
		select {
		case event := <-s.eventChan:
			// Only send events that match the requested subject
			if event.Subject == req.Subject && event.Id > lastID {
				reply := &pb.StreamEventsFromSubjectReply{
					Events: []*pb.Event{event},
				}
				if err := stream.Send(reply); err != nil {
					return fmt.Errorf("failed to send new event: %v", err)
				}
				lastID = event.Id
			}
		case <-ctx.Done():
			// Client disconnected
			return nil
		}
	}
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	dbUri := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPassword, "localhost", "3306", db)
	d, err := sql.Open("mysql", dbUri)
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()
	queries := database.New(d)
	pb.RegisterEventsDBServer(s, &server{
		queries:   queries,
		eventChan: make(chan *pb.Event),
	})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
