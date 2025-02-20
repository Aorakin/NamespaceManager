package interfaces

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/ticket/dtos"
	"github.com/google/uuid"
)

type TicketUsecase interface {
	HandleTicketCallback(models.GliderTicket) error
	GetMyTicket(uuid.UUID, uuid.UUID) ([]dtos.TicketResponse, error)
	SetPayload(*models.GliderTicket) (*dtos.Payload, error)
	ApporveTicket(uuid.UUID) (*models.GliderTicket, error)
	SendTicket(string, interface{}) (int, []map[string]interface{}, error)
}
