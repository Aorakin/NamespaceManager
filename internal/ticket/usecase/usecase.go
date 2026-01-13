package usecase

import (
	"github.com/NamespaceManager/internal/models"
	namespaceInterfaces "github.com/NamespaceManager/internal/namespace/interfaces"
	"github.com/NamespaceManager/internal/ticket/dtos"
	"github.com/NamespaceManager/internal/ticket/interfaces"
	userInterfaces "github.com/NamespaceManager/internal/users/interfaces"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type TicketUsecase struct {
	ticketRepository interfaces.TicketRepository
	namespaceRepo    namespaceInterfaces.NamespaceRepository
	userRepo         userInterfaces.UsersRepository
}

func NewTicketUsecase(ticketRepository interfaces.TicketRepository, namespaceRepo namespaceInterfaces.NamespaceRepository, userRepo userInterfaces.UsersRepository) interfaces.TicketUsecase {
	return &TicketUsecase{ticketRepository: ticketRepository, namespaceRepo: namespaceRepo, userRepo: userRepo}
}

func (u *TicketUsecase) HandleTicketCallback(ticketreq dtos.CreateTicket, userid uuid.UUID) error { //use for test (ใช้จริงคือสร้างจากที่รับมาจาก CH)
	validate := validator.New()
	if err := validate.Struct(ticketreq); err != nil {
		return err
	}
	// ticket := CreateTicketToGliderTicket(ticketreq, userid)
	// if err := u.ticketRepository.Create(ticket); err != nil {
	// 	return err
	// }
	return nil
}

func (u *TicketUsecase) UseTicket(ticketIDs []uuid.UUID) ([]models.GliderTicket, error) {
	tickets := make([]models.GliderTicket, len(ticketIDs))
	// for i, ticketID := range ticketIDs {
	// 	ticket, err := u.ticketRepository.GetTicketByID(ticketID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if ticket.Status != "ready" {
	// 		return nil, fmt.Errorf("ticket not ready")
	// 	}
	// 	if ticket.TaskID != nil && *ticket.TaskID != uuid.Nil {
	// 		return nil, fmt.Errorf("duplicate ticket in other task")
	// 	}
	// 	tickets[i] = ticket
	// 	//ไม่ได้เช็ค ticket owner id match กับ user id
	// }
	return tickets, nil
}

func (u *TicketUsecase) SaveTicket(ticketRes dtos.GliderTicketResponse, name string, ownerID uuid.UUID) error {
	ticket := models.Ticket{
		Name:         name,
		GliderTicket: ticketRes.Ticket,
		Signature:    ticketRes.Signature,
		Status:       models.StatusReady,
		OwnerID:      ownerID,
		TaskID:       nil,
	}

	if err := u.ticketRepository.Create(&ticket); err != nil {
		return err
	}

	return nil
}
