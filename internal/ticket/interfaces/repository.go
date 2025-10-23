package interfaces

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/google/uuid"
)

type TicketRepository interface {
	Create(*models.Ticket) error
	GetTicketByNamespaceID(string) ([]models.Ticket, error)
	GetUserTickets(uuid.UUID) ([]models.Ticket, error)
	UpsertTicketFromCH(*models.Ticket) error
	UpdateTicketStatus(uuid.UUID, models.StatusTicket) error
	GetTicketByGliderTicketID(uuid.UUID) (models.Ticket, error)
	GetTicketNS(uuid.UUID, uuid.UUID) ([]models.GliderTicket, error)
	SendRequest(string, interface{}, string) (int, []byte, error)
	TicketHis(uuid.UUID) ([]models.GliderTicket, error)
	Delete(uuid.UUID) error
	CreateTask(models.Task) error
	GetTasks(uuid.UUID) ([]models.Task, error)
	RemoveTask(uuid.UUID) error
	UpdateTaskStatus(uuid.UUID, models.StatusTicket) error
	GetTasksByID(uuid.UUID) (*models.Task, error)
	ClearTaskID(uuid.UUID) error
	CancelTicket(string) error
}
