package dtos

import "github.com/google/uuid"

type ResetTicketRequest struct {
	TicketIDs []uuid.UUID `json:"ticket_ids" binding:"required"`
}
