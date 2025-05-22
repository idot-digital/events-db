package models

type Event struct {
	ID      int64  `json:"id"`
	Source  string `json:"source"`
	Type    string `json:"type"`
	Subject string `json:"subject"`
	Time    string `json:"time"`
	Data    []byte `json:"data"`
}

type CreateEventRequest struct {
	Source  string `json:"source"`
	Type    string `json:"type"`
	Subject string `json:"subject"`
	Data    []byte `json:"data"`
}

type CreateEventResponse struct {
	ID int64 `json:"id"`
}
