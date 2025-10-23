package dtos

import "github.com/google/uuid"

type CreateTaskRequest struct {
	Title   string      `json:"title"`
	Tickets []uuid.UUID `json:"tickets"`
}
