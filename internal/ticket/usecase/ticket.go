package usecase

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"slices"

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
		log.Printf("failed to unmarshal glider ticket response: %v", err)
		return nil, apiError.NewInternalServerError("Failed to process ticket response from server")
	}

	ticket := models.Ticket{
		Name:            request.Name,
		GliderTicket:    gliderTicket.Ticket,
		GliderTicketID:  gliderTicket.Ticket.ID,
		Signature:       gliderTicket.Signature,
		Status:          models.StatusReady,
		OwnerID:         userID,
		OwnerName:       username,
		NamespaceID:     request.NamespaceID,
		ResourcePoolID:  gliderTicket.Ticket.ResourcePoolID,
		GlideletURN:     gliderTicket.Ticket.GlideletURN,
		NodeDisplayName: gliderTicket.Ticket.NodeDisplayName,
	}

	if err := u.ticketRepository.Create(&ticket); err != nil {
		log.Printf("failed to save ticket: %v", err)
		return nil, apiError.NewInternalServerError("Failed to save ticket, please try again")
	}

	return &gliderTicket, nil
}

func (u *TicketUsecase) GetTicketByNamespaceID(namespaceId uuid.UUID) ([]models.Ticket, error) {
	return u.ticketRepository.GetTicketByNamespaceID(namespaceId)
}

func (u *TicketUsecase) GetTicketByNamespaceIDAndNodeID(namespaceId uuid.UUID, nodeID uuid.UUID) ([]models.Ticket, error) {
	return u.ticketRepository.GetTicketByNamespaceIDAndNodeID(namespaceId, nodeID)
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
		log.Printf("failed to cancel ticket %s: %v", ticketID, err)
		return apiError.NewInternalServerError("Failed to cancel ticket, please try again")
	}
	return nil
}

func (u *TicketUsecase) DeleteTickets(request dtos.DeleteTicketsRequest, userID uuid.UUID) error {
	// validate ownership
	tickets, err := u.ticketRepository.GetTicketsByIDs(request.TicketIDs)
	if err != nil {
		log.Printf("failed to retrieve tickets: %v", err)
		return apiError.NewInternalServerError("Failed to retrieve tickets")
	}

	for _, ticket := range tickets {
		if ticket.OwnerID != userID {
			return apiError.NewForbiddenError("You do not have permission to delete this ticket")
		}

		if !slices.Contains(models.UneditableStatus, ticket.Status) {
			return apiError.NewBadRequestError("One or more tickets cannot be deleted in their current status")
		}
	}

	// proceed to delete
	if err := u.ticketRepository.DeleteTicketsByIDs(request.TicketIDs); err != nil {
		log.Printf("failed to delete tickets: %v", err)
		return apiError.NewInternalServerError("Failed to delete tickets, please try again")
	}

	return nil
}
