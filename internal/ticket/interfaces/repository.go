package interfaces

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/google/uuid"
)

type TicketRepository interface {
	Create(*models.GliderTicket) error
	GetTicketByID(uuid.UUID, uuid.UUID) (models.GliderTicket, error)
	GetTicketNS(uuid.UUID, uuid.UUID) ([]models.GliderTicket, error)
	SendRequest(string, interface{}, string) (int, []byte, error)
	TicketHis(uuid.UUID) ([]models.GliderTicket, error)
	UpdateStatus(uuid.UUID, models.StatusTicket) error
	Delete(uuid.UUID) error
	CreateTask(models.Tasks) error
	GetTasks(uuid.UUID) ([]models.Tasks, error)
	RemoveTasks(uuid.UUID) error
	GetTasksByID(uuid.UUID) (*models.Tasks, error)
}
