package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Task struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey;unique" json:"id"`
	Title              string         `gorm:"type:varchar(255);not null;default:''" json:"title"`
	Tickets            []Ticket       `gorm:"foreignKey:TaskID;" json:"tickets"`
	Status             StatusTicket   `gorm:"type:varchar(50);not null;default:'pending'" json:"status"`
	EstimatedStartTime time.Time      `json:"estimated_start_time"`
	OwnerID            uuid.UUID      `gorm:"type:uuid;not null" json:"owner_id"`
	CreatedAt          time.Time      `json:"created_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	StartedAt          *time.Time     `json:"started_at"`
}

func (t *Task) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}
