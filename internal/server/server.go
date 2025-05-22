package server

import (
	"container/list"
	"fmt"

	"github.com/idot-digital/events-db/database"
	pb "github.com/idot-digital/events-db/grpc"
	"github.com/idot-digital/events-db/internal/models"
)

// Server is used to implement both gRPC and REST servers
type Server struct {
	pb.UnimplementedEventsDBServer
	queries             *database.Queries
	eventEmitterChannel chan *models.Event
	eventListeners      *list.List
	listenerIdCounter   int
}

func New(queries *database.Queries, bufferSize int) *Server {

	emitterChannel := make(chan *models.Event, bufferSize)
	listeners := list.New()

	go func() {
		for event := range emitterChannel {
			for listener := listeners.Front(); listener != nil; listener = listener.Next() {
				listener.Value.(chan *models.Event) <- event
			}
		}
		fmt.Println("Channel closed, reader exiting.")
	}()

	return &Server{
		queries:             queries,
		eventEmitterChannel: emitterChannel,
		eventListeners:      listeners,
	}
}

func (s *Server) GetEmitterChan() chan *models.Event {
	return s.eventEmitterChannel
}

func (s *Server) GetQueries() *database.Queries {
	return s.queries
}

func (s *Server) AttachListener(bufferSize int) (chan *models.Event, *list.Element) {
	s.listenerIdCounter += 1
	channel := make(chan *models.Event, bufferSize)
	elmt := s.eventListeners.PushBack(channel)
	return channel, elmt
}

func (s *Server) DetachListener(listener *list.Element) {
	s.eventListeners.Remove(listener)
}
