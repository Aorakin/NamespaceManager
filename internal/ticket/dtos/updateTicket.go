package dtos

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/google/uuid"
)

type UpdateTicketStatusReq struct {
	ID     uuid.UUID           `json:"id"`
	Status models.StatusTicket `json:"status"`
}
