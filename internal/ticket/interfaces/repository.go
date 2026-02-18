package interfaces

import (
	"time"

	"github.com/NamespaceManager/internal/models"
	"github.com/google/uuid"
)

type TicketRepository interface {
	// ticket
	Create(*models.Ticket) error
	GetTicketsByTaskID(uuid.UUID, bool) ([]models.Ticket, error)
	GetTicketByNamespaceIDAndNodeID(namespaceID uuid.UUID, nodeID uuid.UUID) ([]models.Ticket, error)
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
	UpdateTaskQueueInfo(uuid.UUID, time.Time, models.StatusTicket) error

	GetTicketsByGliderTicketIDs(gliderTicketIDs []uuid.UUID) ([]models.Ticket, error)
	UpdateCodeServerInfo(ticketID uuid.UUID, url string, password string) error
	BatchUpdateTicketStatuses(updates map[uuid.UUID]models.StatusTicket) error
	GetNextQueueTask() (*models.Task, error)
	GetQueuedTasks() ([]models.Task, error)

	// queue
	GetNextStartTime(poolID uuid.UUID, nodeNames []string) (time.Time, error)
	CreateQueueTicket(queueTicket models.QueueTicket) error
	GetHeadTask() (*models.Task, error)
	GetHeadTasksByPoolAndNodes(poolID uuid.UUID, nodeNames []string) (map[string]*models.QueueTicket, error)
	DeleteQueue(taskID uuid.UUID) error
	GetTasksByNodeNames(nodeNames []string) ([]models.Task, error)

	GetTicketsByIDs(ticketIDs []uuid.UUID) ([]models.Ticket, error)
	DeleteTicketsByIDs(ticketIDs []uuid.UUID) error
	GetTasksByIDs(taskIDs []uuid.UUID) ([]models.Task, error)
	DeleteTasksByIDs(taskIDs []uuid.UUID) error
}
