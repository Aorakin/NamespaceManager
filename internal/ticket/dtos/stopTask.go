package dtos

import "github.com/google/uuid"

type StopTaskTickets struct {
	TicketIDs uuid.UUIDs `json:"ticket_ids"`
}
