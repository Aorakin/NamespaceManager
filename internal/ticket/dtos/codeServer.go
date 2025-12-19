package dtos

import "github.com/google/uuid"

type CodeServerResponse struct {
	TicketResponse []TicketResponse `json:"ticket_res"`
}

type TicketResponse struct {
	TicketID uuid.UUID `json:"ticket_id"`
	URL      string    `json:"host_name"`
	Password string    `json:"password"`
}
