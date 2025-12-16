package usecase

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
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

	log.Println("response from CH", string(res))
	log.Printf("%#v", gliderTicket)

	ticket := models.Ticket{
		Name:           request.Name,
		GliderTicket:   gliderTicket.Ticket,
		Signature:      gliderTicket.Signature,
		Status:         models.StatusReady,
		OwnerID:        userID,
		NamespaceID:    request.NamespaceID,
		ResourcePoolID: gliderTicket.Ticket.ResourcePoolID,
		GlideletURN:    gliderTicket.Ticket.GlideletURN,
	}

	if err := u.ticketRepository.Create(&ticket); err != nil {
		return nil, apiError.NewInternalServerError(fmt.Errorf("failed to save ticket: %w", err))
	}

	return &gliderTicket, nil
}

func (u *TicketUsecase) GetTicketByNamespaceID(namespaceId uuid.UUID) ([]models.Ticket, error) {
	return u.ticketRepository.GetTicketByNamespaceID(namespaceId)
}

func (u *TicketUsecase) GetUserTickets(userID uuid.UUID) ([]models.Ticket, error) {
	return u.ticketRepository.GetUserTickets(userID)
}

func (u *TicketUsecase) CancelTicket(ticketID string, accessToken string) error {
	url := os.Getenv("CLEARINGHOUSE_URL") + "/tickets/" + url.PathEscape(ticketID) + "/cancel"
	_, err := httpclient.SendRequestWithAccessToken(url, nil, "PATCH", accessToken)
	if err != nil {
		return nil
	}

	if err := u.ticketRepository.CancelTicket(ticketID); err != nil {
		return apiError.NewInternalServerError(fmt.Errorf("failed to cancel ticket: %w", err))
	}
	return nil
}
