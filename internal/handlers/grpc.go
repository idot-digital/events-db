package handlers

import (
	"context"
	"database/sql"
	"time"

	"github.com/idot-digital/events-db/database"
	pb "github.com/idot-digital/events-db/grpc"
	"github.com/idot-digital/events-db/internal/models"
	"github.com/idot-digital/events-db/internal/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCHandlers implements the gRPC server interface
type GRPCHandlers struct {
	pb.UnimplementedEventsDBServer
	server          *server.Server
	streamBatchSize int32
}

func NewGRPCHandlers(s *server.Server, streamBatchSize int) *GRPCHandlers {
	return &GRPCHandlers{
		server:          s,
		streamBatchSize: int32(streamBatchSize),
	}
}

func (h *GRPCHandlers) CreateEvent(ctx context.Context, req *pb.CreateEventRequest) (*pb.CreateEventReply, error) {
	id, err := h.server.GetQueries().CreateEvent(ctx, database.CreateEventParams{
		Source:  req.Source,
		Type:    req.Type,
		Subject: req.Subject,
		Data:    req.Data,
	})
	if err != nil {
		h.server.GetLogger().Error("Failed to create event", "error", err)
		return nil, status.Error(codes.Internal, "Failed to create event")
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
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "Event not found")
		}
		h.server.GetLogger().Error("Failed to get event", "id", req.Id, "error", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	time, err := res.Time.MarshalText()
	if err != nil {
		h.server.GetLogger().Error("Failed to marshal time", "error", err)
		return nil, status.Error(codes.Internal, "Failed to process event time")
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

func (h *GRPCHandlers) StreamEventsFromSubject(req *pb.EventsFromSubjectRequest, stream pb.EventsDB_StreamEventsFromSubjectServer) error {
	ctx := stream.Context()
	var lastID int64
	if req.FromId != nil {
		lastID = *req.FromId
	}

	for {
		events, err := h.server.GetQueries().GetEventsBySubject(ctx, database.GetEventsBySubjectParams{
			Subject: req.Subject,
			Limit:   h.streamBatchSize,
			ID:      lastID,
		})
		if err != nil {
			h.server.GetLogger().Error("Failed to get events", "subject", req.Subject, "error", err)
			return status.Error(codes.Internal, "Failed to get events")
		}

		if len(events) == 0 {
			// No events found, but this is not an error - just an empty result
			break
		}

		pbEvents := make([]*pb.Event, 0, len(events))
		for _, event := range events {
			time, err := event.Time.MarshalText()
			if err != nil {
				h.server.GetLogger().Error("Failed to marshal time", "error", err)
				return status.Error(codes.Internal, "Failed to process event time")
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
			reply := &pb.EventsFromSubjectReply{
				Events: pbEvents,
			}
			lastID = pbEvents[len(pbEvents)-1].Id
			if err := stream.Send(reply); err != nil {
				return status.Error(codes.Internal, "Failed to send events")
			}
		}
	}

	channel, listener, err := h.server.AttachListener()
	if err != nil {
		h.server.GetLogger().Error("Failed to attach listener", "subject", req.Subject, "error", err)
		return status.Error(codes.ResourceExhausted, "Too many clients for this subject")
	}

	defer h.server.DetachListener(listener)

	for {
		select {
		case event := <-channel:
			if event.Subject == req.Subject && event.ID > lastID {
				reply := &pb.EventsFromSubjectReply{
					Events: []*pb.Event{{
						Id:      event.ID,
						Source:  event.Source,
						Type:    event.Type,
						Subject: event.Subject,
						Time:    event.Time,
						Data:    event.Data,
					}},
					HasMore: true,
				}
				if err := stream.Send(reply); err != nil {
					return status.Error(codes.Internal, "Failed to send new event")
				}
				lastID = event.ID
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (h *GRPCHandlers) GetHistoricEventsFromSubject(ctx context.Context, req *pb.EventsFromSubjectRequest) (*pb.EventsFromSubjectReply, error) {
	var lastID int64
	if req.FromId != nil {
		lastID = *req.FromId
	}

	events, err := h.server.GetQueries().GetEventsBySubject(ctx, database.GetEventsBySubjectParams{
		Subject: req.Subject,
		Limit:   h.streamBatchSize, //TODO: use a different setting
		ID:      lastID,
	})
	if err != nil {
		h.server.GetLogger().Error("Failed to get events", "subject", req.Subject, "error", err)
		return nil, status.Error(codes.Internal, "Failed to get events")
	}

	if len(events) == 0 {
		return &pb.EventsFromSubjectReply{
			Events:  []*pb.Event{},
			HasMore: false,
		}, nil
	}

	pbEvents := make([]*pb.Event, 0, len(events))
	for _, event := range events {
		time, err := event.Time.MarshalText()
		if err != nil {
			h.server.GetLogger().Error("Failed to marshal time", "error", err)
			return nil, status.Error(codes.Internal, "Failed to process event time")
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

	return &pb.EventsFromSubjectReply{
		Events:  pbEvents,
		HasMore: len(events) == int(h.streamBatchSize),
	}, nil
}

func (h *GRPCHandlers) DeleteFromSubject(ctx context.Context, req *pb.EventsFromSubjectRequest) (*pb.OkReply, error) {
	var ID int64
	if req.FromId != nil {
		ID = *req.FromId
	} else {
		ID = 0
	}

	if err := h.server.GetQueries().DeleteFromSubject(ctx, database.DeleteFromSubjectParams{
		Subject: req.Subject,
		ID:      ID,
	}); err != nil {
		h.server.GetLogger().Error("Failed to delete events", "subject", req.Subject, "error", err)
		return nil, status.Error(codes.Internal, "Failed to delete events")
	}

	return &pb.OkReply{
		Ok: true,
	}, nil
}
