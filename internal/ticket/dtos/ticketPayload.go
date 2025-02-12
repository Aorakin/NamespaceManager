package dtos

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/google/uuid"
)

type Payload struct {
	GlideletURN       string              `json:"glidelet_urn"`
	ID                string              `json:"id"`
	Lease             string              `json:"lease"`
	NamespaceURN      string              `json:"namespace_urn"`
	RedeemTimeout     string              `json:"redeem_timeout"`
	ReferenceTicketID string              `json:"reference_ticket_id"`
	Signature         string              `json:"signature"`
	Spec              []models.GliderSpec `json:"spec"`
}

type RequestWithNS struct {
	UserID      uuid.UUID `json:"user_id"`
	NamespaceID uuid.UUID `json:"namespace_id"`
}

type TicketIDRequest struct {
	ID uuid.UUID `json:"id" `
}
