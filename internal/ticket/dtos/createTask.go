package dtos

import "github.com/google/uuid"

type CreateTaskRequest struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Tickets     []uuid.UUID `json:"tickets"`
}
