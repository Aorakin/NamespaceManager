package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SpecResource struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;not null" json:"id"`
	Name     string    `gorm:"not null" json:"name"`
	SpecID   uuid.UUID `gorm:"type:uuid;not null" json:"-"`
	Quantity string    `gorm:"not null" json:"quantity"`
	Unit     string    `gorm:"not null" json:"unit"`
}

func (base *SpecResource) BeforeCreate(tx *gorm.DB) (err error) {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	return
}
