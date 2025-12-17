package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Ticket struct {
	BaseModel
	Name           string       `json:"name"`
	GliderTicket   GliderTicket `gorm:"type:jsonb" json:"ticket"`
	GliderTicketID uuid.UUID    `gorm:"type:uuid;index" json:"glider_ticket_id"`
	Signature      string       `json:"signature"`
	Status         StatusTicket `json:"status"`
	TaskID         *uuid.UUID   `gorm:"type:uuid" json:"task_id"`
	OwnerID        uuid.UUID    `gorm:"type:uuid" json:"owner_id"`
	NamespaceID    uuid.UUID    `gorm:"type:uuid" json:"namespace_id"`
	GlideletURN    string       `json:"glidelet_urn"`
	ResourcePoolID uuid.UUID    `gorm:"type:uuid" json:"resource_pool_id"`
	URL            string       `json:"url"`
	Password       string       `json:"password"`
}

func (t *Ticket) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}
