package models

import "github.com/google/uuid"

type Tasks struct {
	BaseModel
	Owner_ID uuid.UUID      `gorm:"type:uuid;not null" json:"owner_id"`
	Title 		 string         `gorm:"type:varchar(255);not null;default:''" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Tickets  []GliderTicket `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"ticket"`
}
