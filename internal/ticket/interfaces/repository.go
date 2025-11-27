package interfaces

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/google/uuid"
)

type TicketRepository interface {
	// ticket
	Create(*models.Ticket) error
	GetTicketByNamespaceID(string) ([]models.Ticket, error)
	GetUserTickets(uuid.UUID) ([]models.Ticket, error)
	UpdateTicketStatus(uuid.UUID, models.StatusTicket) error
	Update(ticket models.Ticket) error
	GetTicketByGliderTicketID(uuid.UUID) (models.Ticket, error)
	CancelTicket(string) error
	// task
	SendRequest(string, interface{}, string) (int, []byte, error)
	CreateTask(models.Task) error
	GetTasks(uuid.UUID) ([]models.Task, error)
	UpdateTaskStatus(uuid.UUID, models.StatusTicket) error
	GetTasksByID(uuid.UUID) (*models.Task, error)
	ClearTaskID(uuid.UUID) error
}
