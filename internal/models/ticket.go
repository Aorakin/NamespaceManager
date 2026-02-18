package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Ticket struct {
	BaseModel
	Name           string         `json:"name"`
	GliderTicket   GliderTicket   `gorm:"type:jsonb" json:"ticket"`
	GliderTicketID uuid.UUID      `gorm:"type:uuid;index" json:"glider_ticket_id"`
	Signature      string         `json:"signature"`
	Status         StatusTicket   `json:"status"`
	TaskID         *uuid.UUID     `gorm:"type:uuid" json:"task_id"`
	OwnerID        uuid.UUID      `gorm:"type:uuid" json:"owner_id"`
	OwnerName      string         `json:"owner_name"`
	NamespaceID    uuid.UUID      `gorm:"type:uuid" json:"namespace_id"`
	GlideletURN    string         `json:"glidelet_urn"`
	ResourcePoolID uuid.UUID      `gorm:"type:uuid" json:"resource_pool_id"`
	URL            string         `json:"url"`
	Password       string         `json:"password"`
	Failed         bool           `gorm:"index;default:false;not null" json:"failed"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (Ticket) DefaultScope(db *gorm.DB) *gorm.DB {
	return db.Where("failed = ?", false)
}
