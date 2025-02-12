package dtos

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/google/uuid"
)

type TicketResponse struct {
	ID                uuid.UUID           `json:"id"`
	OwnerID           uuid.UUID           `json:"user_id"`
	Spec              []models.GliderSpec `json:"spec" `
	ReferenceTicketID string              `json:"reference_ticket_id"`
	RedeemTimeout     string              `json:"redeem_timeout"`
	Lease             string              `json:"lease"`
	Signature         string              `json:"signature"`
}
