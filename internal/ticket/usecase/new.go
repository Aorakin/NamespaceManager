package usecase

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/ticket/dtos"
	apiError "github.com/NamespaceManager/pkg/api_error"
	"github.com/NamespaceManager/pkg/httpclient"
	"github.com/google/uuid"
)

func (u *TicketUsecase) RequestTicket(request dtos.RequestTicketDTO, accessToken string, userID uuid.UUID) (*dtos.GliderTicketResponse, error) {
	// Call external service to request ticket
	url := os.Getenv("CLEARINGHOUSE_URL") + "/tickets/"
	res, err := httpclient.SendRequestWithAccessToken(url, request, "POST", accessToken)
	if err != nil {
		return nil, err
	}

	var gliderTicket dtos.GliderTicketResponse
	if err := json.Unmarshal(res, &gliderTicket); err != nil {
		return nil, err
	}

	ticket := models.Ticket{
		Name:         request.Name,
		GliderTicket: models.GliderTicketJSON(gliderTicket.Ticket),
		Signature:    gliderTicket.Signature,
		Status:       models.StatusReady,
		OwnerID:      userID,
		NamespaceID:  request.NamespaceID,
		GlideletURN:  gliderTicket.Ticket.GlideletURN,
	}

	if err := u.ticketRepository.Create(&ticket); err != nil {
		return nil, apiError.NewInternalServerError(fmt.Errorf("failed to save ticket: %w", err))
	}

	return &gliderTicket, nil
}

func (u *TicketUsecase) GetTicketByNamespaceID(namespaceId string) ([]models.Ticket, error) {
	return u.ticketRepository.GetTicketByNamespaceID(namespaceId)
}

func (u *TicketUsecase) GetUserTickets(userID uuid.UUID) ([]models.Ticket, error) {
	return u.ticketRepository.GetUserTickets(userID)
}
