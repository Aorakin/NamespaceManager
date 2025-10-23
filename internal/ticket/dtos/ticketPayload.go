package dtos

// import (
// 	"github.com/NamespaceManager/internal/models"
// 	"github.com/google/uuid"
// )

// type Payload struct {
//     ID                string   `json:"id"`
//     NamespaceURN      string       `json:"namespace_urn"`
//     GlideletURN       string       `json:"glidelet_urn"`
//     Spec              []models.GliderSpec `json:"spec"`
//     ReferenceTicketID string       `json:"reference_ticket_id"`
//     RedeemTimeout     string       `json:"redeem_timeout"`
//     Lease             string       `json:"lease"`
//     Signature         string       `json:"signature"`
// }

// type GliderSpec struct {
//     Type      ResourceUnitType `gorm:"not null" json:"type"`
//     PoolID    uuid.UUID        `gorm:"not null" json:"pool_id"`
//     Resources []SpecResource   `gorm:"foreignKey:SpecID" json:"resource" validate:"required,min=1"`
// }

// type SpecResource struct {
//     Name     string `gorm:"not null" json:"name"`
//     Quantity int64  `gorm:"not null" json:"quantity"`
//     Unit     string `gorm:"not null" json:"unit"`
// }

// type RequestWithNS struct {
// 	NamespaceID uuid.UUID `json:"namespace_id"`
// }

// type TicketIDRequest struct {
// 	ID uuid.UUID `json:"id" `
// }
// type ResourceUnitType string
// type ResourceUnitStatus string

// const (
//     ResourceUnitCompute ResourceUnitType = "compute"
//     ResourceUnitStorage ResourceUnitType = "storage"
//     ResourceUnitNetwork ResourceUnitType = "network"
//     ResourceUnitService ResourceUnitType = "service"

//     ResourceUnitAllocated   ResourceUnitStatus = "allocated"
//     ResourceUnitProvisioned ResourceUnitStatus = "provisioned"
//     ResourceUnitUnallocated ResourceUnitStatus = "unallocated"
// )
