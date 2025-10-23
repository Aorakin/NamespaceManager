package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Ticket struct {
	ID           uuid.UUID        `gorm:"type:uuid;primaryKey;unique" json:"id"`
	Name         string           `json:"name"`
	GliderTicket GliderTicketJSON `gorm:"type:jsonb" json:"ticket"`
	Signature    string           `json:"signature"`
	Status       StatusTicket     `json:"status"`
	TaskID       *uuid.UUID       `gorm:"type:uuid" json:"task_id"`
	OwnerID      uuid.UUID        `gorm:"type:uuid" json:"owner_id"`
}

func (t *Ticket) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}

type GliderTicketJSON GliderTicket

func (g GliderTicketJSON) Value() (driver.Value, error) {
	return json.Marshal(g)
}

func (g *GliderTicketJSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	return json.Unmarshal(bytes, g)
}
