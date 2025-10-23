package dtos

import (
	"time"

	"github.com/NamespaceManager/internal/models"
)

type TicketDTO struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Status         string        `json:"status"`
	StartTime      *time.Time    `json:"start_time"`
	EndTime        *time.Time    `json:"end_time"`
	Duration       int           `json:"duration"`
	Price          float64       `json:"price"`
	OwnerID        string        `json:"owner_id"`
	NamespaceID    string        `json:"namespace_id"`
	ResourcePoolID string        `json:"resource_pool_id"`
	QuotaID        string        `json:"quota_id"`
	Resources      []ResourceDTO `json:"resources"`
}

type GliderTicketResponse struct {
	Ticket    models.GliderTicket `json:"ticket"`
	Signature string              `json:"signature"`
}