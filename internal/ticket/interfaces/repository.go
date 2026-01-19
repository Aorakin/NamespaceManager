package interfaces

import (
	"time"

	"github.com/NamespaceManager/internal/models"
	"github.com/google/uuid"
)

type TicketRepository interface {
	// ticket
	Create(*models.Ticket) error
	GetTicketsByTaskID(uuid.UUID) ([]models.Ticket, error)
	GetTicketByNamespaceID(namespaceID uuid.UUID) ([]models.Ticket, error)
	GetUserTickets(userID uuid.UUID) ([]models.Ticket, error)
	UpdateTicketStatus(uuid.UUID, models.StatusTicket) error
	Update(ticket models.Ticket) error
	GetTicketByGliderTicketID(uuid.UUID) (models.Ticket, error)
	CancelTicket(string) error
	// task
	CreateTask(models.Task) error
	GetTasks(uuid.UUID) ([]models.Task, error)
	UpdateTaskStatus(uuid.UUID, models.StatusTicket) error
	GetTasksByID(uuid.UUID) (*models.Task, error)
	ClearTaskID(uuid.UUID) error
	DeleteTask(uuid.UUID) error

	GetTicketsByGliderTicketIDs(gliderTicketIDs []uuid.UUID) ([]models.Ticket, error)
	UpdateCodeServerInfo(ticketID uuid.UUID, url string, password string) error
	BatchUpdateTicketStatuses(updates map[uuid.UUID]models.StatusTicket) error
	GetNextQueueTask() (*models.Task, error)
	GetQueuedTasks() ([]models.Task, error)

	// queue
	GetNextStartTimeByPoolID(poolID uuid.UUID) (time.Time, error)
	CreateQueueTicket(queueTicket models.QueueTicket) error
}
