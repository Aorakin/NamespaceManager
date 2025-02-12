package models

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 1 ticket per glidelet
type GliderTicket struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;unique" json:"id"`
	OwnerID      uuid.UUID `gorm:"type:uuid;not null" json:"owner_id"`
	NamespaceID  uuid.UUID `gorm:"type:uuid;not null;index" json:"namespace_id"`
	NamespaceURN string    `gorm:"not null" json:"namespace_urn"`
	GlideletURN  string    `gorm:"not null" json:"glidelet_urn" `

	Spec []GliderSpec `gorm:"foreignKey:TicketID" json:"spec" validate:"required,min=1" `

	ReferenceTicketID string `gorm:"not null" json:"reference_ticket_id"`

	RedeemTimeout string `gorm:"not null" json:"redeem_timeout" validate:"required"`
	Lease         string `gorm:"not null" json:"lease" validate:"required"`
	Signature     string `gorm:"not null" json:"signature" validate:"required"`
}

func (base *GliderTicket) BeforeCreate(tx *gorm.DB) (err error) {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	if len(base.Spec) == 0 {
		return errors.New("ticket must have at least one spec")
	}
	return
}
