package models

import (
	"time"

	"github.com/google/uuid"
)

type Ticket struct {
	BaseModel
	ExternalID     string `gorm:"uniqueIndex"`
	Name           string
	Status         string
	StartTime      *time.Time
	EndTime        *time.Time
	Duration       int
	Price          float64
	OwnerID        string
	NamespaceID    string
	ResourcePoolID string
	QuotaID        string
	Resources      []Resource `gorm:"foreignKey:TicketID"`
}

type Resource struct {
	BaseModel
	ExternalID string
	Quantity   int
	TicketID   uuid.UUID
}