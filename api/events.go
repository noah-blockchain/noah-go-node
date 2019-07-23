package api

import (
	"github.com/PillarDevelopment/noax-go-node/eventsdb"
	"github.com/PillarDevelopment/noax-go-node/eventsdb/events"
)

type EventsResponse struct {
	Events events.Events `json:"events"`
}

func Events(height uint64) (*EventsResponse, error) {
	return &EventsResponse{
		Events: eventsdb.GetCurrent().LoadEvents(height),
	}, nil
}
