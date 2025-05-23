package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/idot-digital/events-db/database"
	"github.com/idot-digital/events-db/internal/models"
	"github.com/idot-digital/events-db/internal/server"
)

// HTTPHandlers implements the HTTP server handlers
type HTTPHandlers struct {
	server          *server.Server
	streamBatchSize int32
}

func NewHTTPHandlers(s *server.Server, streamBatchSize int) *HTTPHandlers {
	return &HTTPHandlers{
		server:          s,
		streamBatchSize: int32(streamBatchSize),
	}
}

func (h *HTTPHandlers) CreateEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id, err := h.server.GetQueries().CreateEvent(r.Context(), database.CreateEventParams{
		Source:  req.Source,
		Type:    req.Type,
		Subject: req.Subject,
		Data:    req.Data,
	})
	if err != nil {
		h.server.GetLogger().Error("Failed to create event", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.CreateEventResponse{ID: id})
}

func (h *HTTPHandlers) GetEventByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	res, err := h.server.GetQueries().GetEventByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}
		h.server.GetLogger().Error("Failed to get event", "id", id, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	time, err := res.Time.MarshalText()
	if err != nil {
		h.server.GetLogger().Error("Failed to marshal time", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	event := models.Event{
		ID:      res.ID,
		Source:  res.Source,
		Type:    res.Type,
		Subject: res.Subject,
		Time:    string(time),
		Data:    res.Data,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

func (h *HTTPHandlers) StreamEventsFromSubjectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	subject := r.URL.Query().Get("subject")
	if subject == "" {
		http.Error(w, "Missing subject parameter", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clientGone := w.(http.CloseNotifier).CloseNotify()
	lastID := int64(0)

	for {
		events, err := h.server.GetQueries().GetEventsBySubject(r.Context(), database.GetEventsBySubjectParams{
			Subject: subject,
			Limit:   h.streamBatchSize,
			ID:      lastID,
		})
		if err != nil {
			h.server.GetLogger().Error("Failed to get events", "subject", subject, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if len(events) == 0 {
			// No events found, but this is not an error - just an empty result
			break
		}

		for _, event := range events {
			time, err := event.Time.MarshalText()
			if err != nil {
				h.server.GetLogger().Error("Failed to marshal time", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			eventJSON, err := json.Marshal(models.Event{
				ID:      event.ID,
				Source:  event.Source,
				Type:    event.Type,
				Subject: event.Subject,
				Time:    string(time),
				Data:    event.Data,
			})
			if err != nil {
				h.server.GetLogger().Error("Failed to marshal event", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			fmt.Fprintf(w, "data: %s\n\n", eventJSON)
			w.(http.Flusher).Flush()
			lastID = event.ID
		}
	}

	channel, listener, err := h.server.AttachListener()
	if err != nil {
		h.server.GetLogger().Error("Failed to attach listener", "subject", subject, "error", err)
		http.Error(w, "Too many clients for this subject", http.StatusTooManyRequests)
		return
	}

	defer h.server.DetachListener(listener)
	for {
		select {
		case event := <-channel:
			if event.Subject == subject && event.ID > lastID {
				eventJSON, err := json.Marshal(event)
				if err != nil {
					h.server.GetLogger().Error("Failed to marshal event", "error", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}

				fmt.Fprintf(w, "data: %s\n\n", eventJSON)
				w.(http.Flusher).Flush()
				lastID = event.ID
			}
		case <-clientGone:
			return
		}
	}
}

func (h *HTTPHandlers) GetSubjectsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	subjects, err := h.server.GetQueries().GetAvailableSubjects(r.Context())
	if err != nil {
		h.server.GetLogger().Error("Failed to get subjects", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subjects)
}
