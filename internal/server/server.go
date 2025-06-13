package server

import (
	"container/list"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/idot-digital/events-db/database"
	pb "github.com/idot-digital/events-db/grpc"
	"github.com/idot-digital/events-db/internal/metrics"
	"github.com/idot-digital/events-db/internal/models"
)

// EventFilter defines criteria for filtering events
type EventFilter struct {
	Subject   string
	Type      *string // Optional type filter
	Recursive bool    // Whether to match subject prefixes
}

// FilteredListener represents a listener with its filter criteria
type FilteredListener struct {
	Channel chan *models.Event
	Filter  EventFilter
}

// MatchesEvent checks if an event matches the filter criteria
func (f *EventFilter) MatchesEvent(event *models.Event) bool {
	// Special case: empty subject matches all events (for backward compatibility)
	if f.Subject == "" {
		// Check type matching if specified
		if f.Type != nil && event.Type != *f.Type {
			return false
		}
		return true
	}
	
	// Check subject matching
	subjectMatches := false
	if f.Recursive {
		// Recursive: check if event subject starts with filter subject
		subjectMatches = strings.HasPrefix(event.Subject, f.Subject)
	} else {
		// Exact match
		subjectMatches = event.Subject == f.Subject
	}
	
	if !subjectMatches {
		return false
	}
	
	// Check type matching if specified
	if f.Type != nil && event.Type != *f.Type {
		return false
	}
	
	return true
}

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
				filteredListener := listener.Value.(*FilteredListener)
				if filteredListener.Filter.MatchesEvent(event) {
					select {
					case filteredListener.Channel <- event:
					default:
						// Channel is full, skip this event to prevent blocking
						logger.Warn("Client channel full, dropping event", "subject", event.Subject, "event_id", event.ID)
					}
				}
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
	// For backward compatibility, create a filter that matches all events
	return s.AttachFilteredListener(EventFilter{
		Subject:   "",
		Type:      nil,
		Recursive: true, // Match all subjects
	})
}

func (s *Server) AttachFilteredListener(filter EventFilter) (chan *models.Event, *list.Element, error) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	if s.totalClients >= s.maxTotalClients {
		return nil, nil, fmt.Errorf("maximum number of total clients reached")
	}

	s.listenerIdCounter += 1
	channel := make(chan *models.Event, s.clientBufferSize)
	filteredListener := &FilteredListener{
		Channel: channel,
		Filter:  filter,
	}
	elmt := s.eventListeners.PushBack(filteredListener)
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
