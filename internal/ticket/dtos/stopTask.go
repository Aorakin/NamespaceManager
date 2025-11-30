package dtos

import "github.com/google/uuid"

type StopTaskTickets struct {
	TicketIDs uuid.UUIDs `json:"ticket_ids"`
}

// {data: {task_id: "24f14b58-cd81-45aa-8221-08158dc37436"}}
type StopTaskRequest struct {
	Data TaskID `json:"data"`
}

type TaskID struct {
	TaskID uuid.UUID `json:"task_id"`
}

type StopTaskResponse struct {
	TicketID uuid.UUID `json:"ticket_id"`
	Status   string    `json:"status"`
}
