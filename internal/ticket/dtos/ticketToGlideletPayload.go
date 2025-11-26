package dtos

import "github.com/google/uuid"

type TicketReq struct {
	GlideletURN       string     `json:"glidelet_urn"`
	ID                uuid.UUID  `json:"id"`
	Lease             string     `json:"lease"`
	NamespaceURN      string     `json:"namespace_urn"`
	RedeemTimeout     string     `json:"redeem_timeout"`
	ReferenceTicketID string     `json:"reference_ticket_id"`
	Signature         string     `json:"signature"`
	Spec              GliderSpec `json:"spec"`
}

type GliderSpec struct {
	Type      ResourceUnitType `gorm:"not null" json:"type"`
	PoolID    uuid.UUID        `gorm:"not null" json:"pool_id"`
	Resources []SpecResource   `gorm:"foreignKey:SpecID" json:"resource" validate:"required,min=1"`
}

type SpecResource struct {
	Name     string `gorm:"not null" json:"name"`
	Quantity int64  `gorm:"not null" json:"quantity"`
	Unit     string `gorm:"not null" json:"unit"`
}

type ResourceUnitType string

const (
	ResourceUnitCompute ResourceUnitType = "compute"
	ResourceUnitStorage ResourceUnitType = "storage"
	ResourceUnitNetwork ResourceUnitType = "network"
	ResourceUnitService ResourceUnitType = "service"
)
