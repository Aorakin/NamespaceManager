package interfaces

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/ticket/dtos"
	"github.com/google/uuid"
)

type TicketUsecase interface {
	HandleTicketCallback(dtos.CreateTicket, uuid.UUID) error
	UseTicket([]uuid.UUID) ([]models.GliderTicket, error)

	GetTicketByNamespaceID(namespaceID uuid.UUID) ([]models.Ticket, error)
	GetUserTickets(uuid.UUID) ([]models.Ticket, error)
	SaveTicket(dtos.GliderTicketResponse, string, uuid.UUID) error
	UpdateTicketStatusFromGlidelet([]dtos.StatusRes) error
	CancelTicket(ticketID string, accessToken string) error

	StopTask(userID uuid.UUID, taskID uuid.UUID) (interface{}, error)
	CancelTask(userID uuid.UUID, taskID uuid.UUID) error

	// new
	RequestTicket(dtos.RequestTicketDTO, string, uuid.UUID, string) (*dtos.GliderTicketResponse, error)
	CreateTask(request *dtos.CreateTaskRequest, userID uuid.UUID) error
	GetTasks(userID uuid.UUID) ([]models.Task, error)

	
}
