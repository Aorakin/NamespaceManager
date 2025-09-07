package interfaces

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/ticket/dtos"
	"github.com/google/uuid"
)

type TicketUsecase interface {
	HandleTicketCallback(dtos.CreateTicket, uuid.UUID) error
	GetTicketNS(uuid.UUID, uuid.UUID) ([]dtos.TicketResponse, error)
	ApporveTicket(uuid.UUID) (models.GliderTicket, error)
	SendTicket([]uuid.UUID) (int, map[string]interface{}, error)
	SetPayload(models.GliderTicket) (*dtos.Payload, error)
	TicketHis(uuid.UUID) ([]dtos.TicketResponse, error)
	UseTicket([]uuid.UUID) ([]models.GliderTicket, error)
	RollbackFailedTickets([]dtos.Payload, int) error
	UpdateStatus(uuid.UUID, models.StatusTicket) error
	Delete(uuid.UUID) error
	CreateTask(dtos.CreateTaskRequest, uuid.UUID) error
	GetTasks(uuid.UUID) ([]models.Tasks, error)
	StopTasks(uuid.UUID) error
}
