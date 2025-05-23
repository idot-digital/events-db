package server

import (
	"container/list"
	"fmt"
	"log/slog"
	"sync"

	"github.com/idot-digital/events-db/database"
	pb "github.com/idot-digital/events-db/grpc"
	"github.com/idot-digital/events-db/internal/metrics"
	"github.com/idot-digital/events-db/internal/models"
)

// Server is used to implement both gRPC and REST servers
type Server struct {
	pb.UnimplementedEventsDBServer
	queries             *database.Queries
	eventEmitterChannel chan *models.Event
	eventListeners      *list.List
	listenerIdCounter   int
	logger              *slog.Logger
	totalClients        int
	clientsMutex        sync.Mutex
	maxTotalClients     int
	clientBufferSize    int
}

func New(queries *database.Queries, bufferSize int, maxTotalClients int, clientBufferSize int, logger *slog.Logger) *Server {
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
		listenerIdCounter:   0,
		logger:              logger,
		totalClients:        0,
		maxTotalClients:     maxTotalClients,
		clientBufferSize:    clientBufferSize,
	}
}

func (s *Server) GetEmitterChan() chan *models.Event {
	return s.eventEmitterChannel
}

func (s *Server) GetQueries() *database.Queries {
	return s.queries
}

func (s *Server) AttachListener() (chan *models.Event, *list.Element, error) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	if s.totalClients >= s.maxTotalClients {
		return nil, nil, fmt.Errorf("maximum number of total clients reached")
	}

	s.listenerIdCounter += 1
	channel := make(chan *models.Event, s.clientBufferSize)
	elmt := s.eventListeners.PushBack(channel)
	s.totalClients++

	// Update active streams metric
	metrics.ActiveEventStreams.Inc()

	return channel, elmt, nil
}

func (s *Server) DetachListener(listener *list.Element) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	s.eventListeners.Remove(listener)
	s.totalClients--

	// Update active streams metric
	metrics.ActiveEventStreams.Dec()
}

func (s *Server) GetLogger() *slog.Logger {
	return s.logger
}
