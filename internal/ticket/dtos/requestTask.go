package dtos

import "github.com/google/uuid"

type RequestTaskID struct {
	TaskID uuid.UUID `json:"task_id"`
}
