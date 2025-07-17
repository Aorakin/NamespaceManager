package dtos

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/google/uuid"
)

type CreateTicket struct {
	NamespaceID  uuid.UUID `json:"namespace_id" validate:"required"`
	NamespaceURN string    `json:"namespace_urn" validate:"required"`
	GlideletURN  string    `json:"glidelet_urn" validate:"required"`

	Spec              []SpecListReq `json:"spec" validate:"required"`
	ReferenceTicketID string        `json:"reference_ticket_id"`

	RedeemTimeout string `json:"redeem_timeout" validate:"required"`
	Lease         string `json:"lease" validate:"required"`
	Signature     string `json:"signature" validate:"required"`
}

type SpecListReq struct {
	Type      models.ResourceUnitType `json:"type" validate:"required"`
	PoolID    uuid.UUID               `json:"pool_id" validate:"required"`
	Resources []SpecResourceReq       `json:"resource" validate:"required"`
}

type SpecResourceReq struct {
	Name     string `json:"name" validate:"required"`
	Quantity string `json:"quantity" validate:"required"`
	Unit     string `json:"unit" validate:"required"`
}
