package models

import (
	"time"

	"github.com/google/uuid"
)

type QueueTicket struct {
	BaseModel
	GliderTicketID uuid.UUID `gorm:"type:uuid;index" json:"glider_ticket_id"`
	PoolID         uuid.UUID `gorm:"type:uuid;index" json:"pool_id"`
	StartTime      time.Time `json:"start_time"`
}
