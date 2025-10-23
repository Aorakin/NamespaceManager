package dtos

type ResourceDTO struct {
	ID       string `json:"resource_id"`
	Quantity int    `json:"quantity"`
}
type RequestTicketDTO struct {
	Name        string        `json:"name" binding:"required"`
	NamespaceID string        `json:"namespace_id" binding:"required"`
	QuotaID     string        `json:"quota_id" binding:"required"`
	Resources   []ResourceDTO `json:"resources" binding:"required"`
	Duration    int           `json:"duration" binding:"required"`
}
