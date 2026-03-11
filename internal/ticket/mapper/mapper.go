package mapper

import (
	"fmt"

	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/ticket/dtos"
	"github.com/google/uuid"
)

// ToUserTicketResponse converts a single ticket model to DTO
func ToUserTicketResponse(ticket models.Ticket) dtos.UserTicketResponse {
	return dtos.UserTicketResponse{
		ID:           ticket.ID.String(),
		Name:         ticket.Name,
		GliderTicket: models.GliderTicket(ticket.GliderTicket),
		Status:       string(ticket.Status),
		OwnerName:    ticket.OwnerName,
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

func ToTicketRequest(ticket models.Ticket) (*dtos.TicketReq, error) {
	// Convert poolID (string) to uuid.UUID
	poolID, err := uuid.Parse(ticket.GliderTicket.Spec.PoolID)
	if err != nil {
		return nil, fmt.Errorf("invalid pool_id: %w", err)
	}

	// Convert resources
	var resources []dtos.SpecResource
	for _, r := range ticket.GliderTicket.Spec.Resources {
		resources = append(resources, dtos.SpecResource{
			Name:     r.Name,
			Quantity: int64(r.Quantity),
			Unit:     r.Unit,
		})
	}

	req := &dtos.TicketReq{
		GlideletURN:       ticket.GliderTicket.GlideletURN,
		ID:                ticket.GliderTicket.ID,
		Lease:             fmt.Sprintf("%d", ticket.GliderTicket.Lease),
		NamespaceURN:      ticket.GliderTicket.NamespaceID.String(),
		RedeemTimeout:     fmt.Sprintf("%d", ticket.GliderTicket.RedeemTimeout),
		ReferenceTicketID: ticket.GliderTicket.ReferenceTicketID.String(),
		NodeName:          ticket.GliderTicket.NodeName,
		Signature:         ticket.Signature,
		Spec: dtos.GliderSpec{
			Type:      dtos.ResourceUnitType(ticket.GliderTicket.Spec.Type),
			PoolID:    poolID,
			Resources: resources,
		},
		Ticket: ticket.GliderTicket,
	}

	return req, nil
}
