package interfaces

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/google/uuid"
)

type TicketRepository interface {
	Create(models.GliderTicket) error
	GetTicketByID(uuid.UUID) (*models.GliderTicket, error)
	GetMyTicket(uuid.UUID, uuid.UUID) ([]models.GliderTicket, error)
	SendRequest(string, interface{}, string) (int, []byte, error)
}
