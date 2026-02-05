package usecase

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/ticket/dtos"
	apiError "github.com/NamespaceManager/pkg/api_error"
	"github.com/NamespaceManager/pkg/httpclient"
	"github.com/google/uuid"
)

func (u *TicketUsecase) RequestTicket(request dtos.RequestTicketDTO, accessToken string, userID uuid.UUID, username string) (*dtos.GliderTicketResponse, error) {
	// Call external service to request ticket
	url := os.Getenv("CLEARINGHOUSE_URL") + "/tickets/"
	res, err := httpclient.SendRequestWithAccessToken(url, request, "POST", accessToken)
	if err != nil {
		return nil, err
	}

	var gliderTicket dtos.GliderTicketResponse
	if err := json.Unmarshal(res, &gliderTicket); err != nil {
		return nil, apiError.NewInternalServerError(fmt.Errorf("failed to unmarshal glider ticket response: %w", err))
	}

	ticket := models.Ticket{
		Name:           request.Name,
		GliderTicket:   gliderTicket.Ticket,
		GliderTicketID: gliderTicket.Ticket.ID,
		Signature:      gliderTicket.Signature,
		Status:         models.StatusReady,
		OwnerID:        userID,
		OwnerName:      username,
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

func (u *TicketUsecase) DeleteTickets(request dtos.DeleteTicketsRequest, userID uuid.UUID) error {
	// validate ownership
	tickets, err := u.ticketRepository.GetTicketsByIDs(request.TicketIDs)
	if err != nil {
		return apiError.NewInternalServerError(fmt.Errorf("failed to retrieve tickets: %w", err))
	}

	for _, ticket := range tickets {
		if ticket.OwnerID != userID {
			return apiError.NewForbiddenError(fmt.Errorf("ticket %s does not belong to the user", ticket.ID))
		}
		if ticket.Status != models.StatusReady && ticket.Status != models.StatusCancelled && ticket.Status != models.StatusStopped {
			return apiError.NewBadRequestError(fmt.Errorf("ticket %s cannot be deleted in its current status", ticket.ID))
		}
	}

	// proceed to delete
	if err := u.ticketRepository.DeleteTicketsByIDs(request.TicketIDs); err != nil {
		return apiError.NewInternalServerError(fmt.Errorf("failed to delete tickets: %w", err))
	}

	return nil
}
