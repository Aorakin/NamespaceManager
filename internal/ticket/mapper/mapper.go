package mapper

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/ticket/dtos"
)

// ToUserTicketResponse converts a single ticket model to DTO
func ToUserTicketResponse(ticket models.Ticket) dtos.UserTicketResponse {
	return dtos.UserTicketResponse{
		ID:           ticket.ID.String(),
		Name:         ticket.Name,
		GliderTicket: models.GliderTicket(ticket.GliderTicket),
		Signature:    ticket.Signature,
		Status:       string(ticket.Status),
	}
}

// ToUserTicketResponseList converts a list of ticket models to DTOs
func ToUserTicketResponseList(tickets []models.Ticket) []dtos.UserTicketResponse {
	dtosList := make([]dtos.UserTicketResponse, 0, len(tickets))
	for _, ticket := range tickets {
		dtosList = append(dtosList, ToUserTicketResponse(ticket))
	}
	return dtosList
}
