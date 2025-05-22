package handlers

import (
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
	server *server.Server
}

func NewHTTPHandlers(s *server.Server) *HTTPHandlers {
	return &HTTPHandlers{server: s}
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
		http.Error(w, fmt.Sprintf("Failed to create event: %v", err), http.StatusInternalServerError)
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
		fmt.Printf("Failed to get event: %v", err) //TODO: do proper logging
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	time, err := res.Time.MarshalText()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal time: %v", err), http.StatusInternalServerError)
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
			Limit:   10, // TODO: Move to config
			ID:      lastID,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get events: %v", err), http.StatusInternalServerError)
			return
		}

		for _, event := range events {
			time, err := event.Time.MarshalText()
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to marshal time: %v", err), http.StatusInternalServerError)
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
				http.Error(w, fmt.Sprintf("Failed to marshal event: %v", err), http.StatusInternalServerError)
				return
			}

			fmt.Fprintf(w, "data: %s\n\n", eventJSON)
			w.(http.Flusher).Flush()
			lastID = event.ID
		}

		if len(events) == 0 {
			break
		}
	}

	channel, listener := h.server.AttachListener(10) //TODO: make this configurable or find a smart solution

	for {
		select {
		case event := <-channel:
			if event.Subject == subject && event.ID > lastID {
				eventJSON, err := json.Marshal(event)
				if err != nil {
					http.Error(w, fmt.Sprintf("Failed to marshal event: %v", err), http.StatusInternalServerError)
					return
				}

				fmt.Fprintf(w, "data: %s\n\n", eventJSON)
				w.(http.Flusher).Flush()
				lastID = event.ID
			}
		case <-clientGone:
			h.server.DetachListener(listener)
			return
		}
	}
}
