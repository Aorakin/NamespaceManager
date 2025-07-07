package models

import "github.com/google/uuid"

type Tasks struct {
	BaseModel
	Owner_ID uuid.UUID      `gorm:"type:uuid;not null" json:"owner_id"`
	Tickets  []GliderTicket `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"ticket"`
}
