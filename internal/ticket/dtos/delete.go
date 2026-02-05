package dtos

import "github.com/google/uuid"

type DeleteTicketsRequest struct {
	TicketIDs []uuid.UUID `json:"ticket_ids" binding:"required"`
}

type DeleteTasksRequest struct {
	TaskIDs []uuid.UUID `json:"task_ids" binding:"required"`
}
