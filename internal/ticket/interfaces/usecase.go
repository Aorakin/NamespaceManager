package interfaces

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/ticket/dtos"
	"github.com/google/uuid"
)

type TicketUsecase interface {
	HandleTicketCallback(dtos.CreateTicket, uuid.UUID) error
	ApproveTicket(uuid.UUID) (dtos.TicketReq, error)
	SendTicket([]uuid.UUID) (int, *dtos.CodeServerResponse, error)
	UseTicket([]uuid.UUID) ([]models.GliderTicket, error)
	CreateTask(dtos.CreateTaskRequest, uuid.UUID) error
	GetTasks(uuid.UUID) ([]models.Task, error)
	StopTask(uuid.UUID) error

	// GetTicketFromCH(string) (int, dtos.GliderTicketResponse, error)
	// RequestTicketToCH(dtos.RequestTicketDTO) (int, dtos.GliderTicketResponse, error)
	GetTicketByNamespaceID(string) ([]dtos.UserTicketResponse, error)
	GetUserTickets(uuid.UUID) ([]dtos.UserTicketResponse, error)
	SaveTicket(dtos.GliderTicketResponse, string, uuid.UUID) error
	UpdateTicketStatusFromGlidelet([]dtos.StatusRes) error
	CancelTicket(string) error
	ConvertTicketToTicketRequest(models.Ticket) (dtos.TicketReq, error)

	GetStopTaskPayload(userID uuid.UUID, taskID uuid.UUID) (uuid.UUIDs, error)
}
