package models

import "github.com/google/uuid"

type ResourceUnit struct {
	ID                 uuid.UUID
	URN                string
	GlideletURN        string
	NamespaceURN       string
	LocalNamespaceID   uuid.UUID
	TicketID           uuid.UUID
	Type               ResourceUnitType   // Compute, Storage, Networking, Service
	Lease              LeaseTimeInterface // 10Hrs, 1Hrs, unlimited, on-demand
	Status             ResourceUnitStatus
	ResourceMetadataID uuid.UUID
	// Identity
	OwnerURN string
}

type ResourceUnitInterface interface {
	Create(ticket GliderTicket) ([]ResourceUnit, error)
	Get(id string) (*ResourceUnit, error)
	List() []ResourceUnit
	Update(ticket GliderTicket) ([]ResourceUnit, error)
	Delete(ticket GliderTicket) error
}

type ResourceInterface interface {
	Create()
	Get()
	Update()
	Delete()
}

type LeaseTimeInterface interface {
	Set()
	Stop()
	Continue()
	Update()
}

type ResourceUnitType string
type ResourceUnitStatus string

const (
	ResourceUnitCompute ResourceUnitType = "compute"
	ResourceUnitStorage ResourceUnitType = "storage"
	ResourceUnitNetwork ResourceUnitType = "network"
	ResourceUnitService ResourceUnitType = "service"

	ResourceUnitAllocated   ResourceUnitStatus = "allocated"
	ResourceUnitProvisioned ResourceUnitStatus = "provisioned"
	ResourceUnitUnallocated ResourceUnitStatus = "unallocated"
)
