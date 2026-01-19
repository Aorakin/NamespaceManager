package dtos

import (
	"time"
)

type QueuePayload struct {
	Tickets []TicketReq `json:"tickets"`
	EndTime time.Time   `json:"end_time"`
}

type PoolQueueResponse struct {
	StartTime time.Time `json:"start_time"`
}
