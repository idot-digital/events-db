package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/idot-digital/events-db/database"
	pb "github.com/idot-digital/events-db/grpc"
	"github.com/idot-digital/events-db/internal/models"
	"github.com/idot-digital/events-db/internal/server"
)

// GRPCHandlers implements the gRPC server interface
type GRPCHandlers struct {
	pb.UnimplementedEventsDBServer
	server *server.Server
}

func NewGRPCHandlers(s *server.Server) *GRPCHandlers {
	return &GRPCHandlers{server: s}
}

func (h *GRPCHandlers) CreateEvent(ctx context.Context, req *pb.CreateEventRequest) (*pb.CreateEventReply, error) {
	id, err := h.server.GetQueries().CreateEvent(ctx, database.CreateEventParams{
		Source:  req.Source,
		Type:    req.Type,
		Subject: req.Subject,
		Data:    req.Data,
	})
	if err != nil {
		return nil, err
	}

	event := &models.Event{
		ID:      id,
		Source:  req.Source,
		Type:    req.Type,
		Subject: req.Subject,
		Time:    time.Now().Format(time.RFC3339),
		Data:    req.Data,
	}

	h.server.GetEmitterChan() <- event

	return &pb.CreateEventReply{
		Id: id,
	}, nil
}

func (h *GRPCHandlers) GetEventByID(ctx context.Context, req *pb.GetEventByIDRequest) (*pb.Event, error) {
	res, err := h.server.GetQueries().GetEventByID(ctx, req.Id)
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

func (h *GRPCHandlers) StreamEventsFromSubject(req *pb.StreamEventsFromSubjectRequest, stream pb.EventsDB_StreamEventsFromSubjectServer) error {
	ctx := stream.Context()
	lastID := int64(0)

	for {
		events, err := h.server.GetQueries().GetEventsBySubject(ctx, database.GetEventsBySubjectParams{
			Subject: req.Subject,
			Limit:   10, // TODO: Move to config
			ID:      lastID,
		})
		if err != nil {
			return fmt.Errorf("failed to get events: %v", err)
		}

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

	channel, listener := h.server.AttachListener(10) //TODO: make this configurable or find a smart solution

	for {
		select {
		case event := <-channel:
			if event.Subject == req.Subject && event.ID > lastID {
				reply := &pb.StreamEventsFromSubjectReply{
					Events: []*pb.Event{{
						Id:      event.ID,
						Source:  event.Source,
						Type:    event.Type,
						Subject: event.Subject,
						Time:    event.Time,
						Data:    event.Data,
					}},
				}
				if err := stream.Send(reply); err != nil {
					return fmt.Errorf("failed to send new event: %v", err)
				}
				lastID = event.ID
			}
		case <-ctx.Done():
			h.server.DetachListener(listener)
			return nil
		}
	}
}
