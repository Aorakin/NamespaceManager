package usecase

import (
	"encoding/json"
	"fmt"

	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/ticket/dtos"
	"github.com/NamespaceManager/internal/ticket/interfaces"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type TicketUsecase struct {
	TicketRepository interfaces.TicketRepository
}

func NewTicketUsecase(ticketRepository interfaces.TicketRepository) interfaces.TicketUsecase {
	return &TicketUsecase{TicketRepository: ticketRepository}
}

func (u *TicketUsecase) HandleTicketCallback(ticket models.GliderTicket) error {
	validate := validator.New()
	if err := validate.Struct(ticket); err != nil {
		return err
	}

	if err := u.TicketRepository.Create(ticket); err != nil {
		return err
	}
	return nil
}

func (u *TicketUsecase) GetMyTicket(userID uuid.UUID, namespaceID uuid.UUID) ([]dtos.TicketResponse, error) {
	tickets, err := u.TicketRepository.GetMyTicket(userID, namespaceID)
	if err != nil {
		return nil, err
	}
	ticketResponses := make([]dtos.TicketResponse, len(tickets))
	for i, ticket := range tickets {
		ticketResponses[i] = dtos.TicketResponse{
			ID:                ticket.ID,
			OwnerID:           ticket.OwnerID,
			Spec:              ticket.Spec,
			ReferenceTicketID: ticket.ReferenceTicketID,
			RedeemTimeout:     ticket.RedeemTimeout,
			Lease:             ticket.Lease,
			Signature:         ticket.Signature,
		}
	}
	return ticketResponses, nil
}

func (u *TicketUsecase) SetPayload(ticket *models.GliderTicket) (*dtos.Payload, error) {
	payload := dtos.Payload{
		GlideletURN:       ticket.GlideletURN,
		ID:                ticket.ID.String(),
		Lease:             ticket.Lease,
		NamespaceURN:      ticket.NamespaceURN,
		RedeemTimeout:     ticket.RedeemTimeout,
		ReferenceTicketID: ticket.ReferenceTicketID,
		Signature:         ticket.ReferenceTicketID,
		Spec:              ticket.Spec,
	}
	return &payload, nil
}

func (u *TicketUsecase) ApporveTicket(ticketID uuid.UUID) (*models.GliderTicket, error) {
	ticket, err := u.TicketRepository.GetTicketByID(ticketID)
	if err != nil {
		return nil, err
	}
	//check ticket with passport
	return ticket, nil
}

func (u *TicketUsecase) SendTicket(url string, payload interface{}) (int, []map[string]interface{}, error) {
	status, body, err := u.TicketRepository.SendRequest(url, payload, "POST")
	if err != nil {
		return 0, nil, err
	}
	var jsonResponse []map[string]interface{}
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return 0, nil, fmt.Errorf("error : Failed to parse response")
	}
	return status, jsonResponse, nil
}
