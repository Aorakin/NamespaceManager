package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StatusTicket string

const (
	Ready    StatusTicket = "ready"
	Pending  StatusTicket = "pending"
	Active   StatusTicket = "active"
	Inactive StatusTicket = "inactive"
)

type GliderTicket struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;unique" json:"id"`
	OwnerID      uuid.UUID  `gorm:"type:uuid;not null" json:"owner_id"`
	NamespaceID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"namespace_id"`
	TaskID       *uuid.UUID `gorm:"type:uuid" json:"task_id"`
	NamespaceURN string     `gorm:"not null" json:"namespace_urn"`
	GlideletURN  string     `gorm:"not null" json:"glidelet_urn" `

	Spec []GliderSpec `gorm:"foreignKey:TicketID" json:"spec" validate:"required,min=1" `

	ReferenceTicketID string `json:"reference_ticket_id"`

	RedeemTimeout string       `gorm:"not null" json:"redeem_timeout" validate:"required"`
	Lease         string       `gorm:"not null" json:"lease" validate:"required"` //unit sec
	Signature     string       `gorm:"not null" json:"signature" validate:"required"`
	Status        StatusTicket `gorm:"not null;default:'ready'" json:"status" `
	CreatedAt     time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
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
