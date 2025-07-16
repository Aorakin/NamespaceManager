package models

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GliderSpec struct {
	ID        uuid.UUID        `gorm:"type:uuid;primaryKey;not null" json:"id"`
	TicketID  uuid.UUID        `gorm:"type:uuid;not null" json:"-"`
	Type      ResourceUnitType `gorm:"not null" json:"type"`
	PoolID    uuid.UUID        `gorm:"not null" json:"pool_id"`
	Resources []SpecResource   `gorm:"foreignKey:SpecID" json:"resource" validate:"required,min=1"`
}

func (base *GliderSpec) BeforeCreate(tx *gorm.DB) (err error) {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	if len(base.Resources) == 0 {
		return errors.New("ticket must have at least one spec")
	}
	return
}
