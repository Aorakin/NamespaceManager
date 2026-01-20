package dtos

import (
	"time"
)

type QueuePayload struct {
	Tickets   []TicketReq `json:"tickets"`
	EndTime   time.Time   `json:"end_time"`
	StartTime time.Time   `json:"start_time"`
}

type PoolQueueResponse struct {
	StartTime time.Time `json:"start_time"`
}

type PoolBackfillResponse struct {
	Status bool `json:"status"`
}
