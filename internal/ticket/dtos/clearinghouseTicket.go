package dtos

import "time"

type ResourceDTO struct {
	ID       string `json:"resource_id"`
	Quantity int    `json:"quantity"`
}

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

type RequestTicketDTO struct {
	Name       	string 				`json:"name" binding:"required"`
	NamespaceID string 				`json:"namespace_id" binding:"required"`
	QuotaID			string 				`json:"quota_id" binding:"required"`
	Resources   []ResourceDTO `json:"resources" binding:"required"`
	Duration   	int    				`json:"duration" binding:"required"`
}