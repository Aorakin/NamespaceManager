package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type GliderTicket struct {
	ID                uuid.UUID  `json:"id"`           // ticket id
	NamespaceID       uuid.UUID  `json:"namespace_id"` // namespace id
	NamespaceName     string     `json:"namespace_name"`
	ResourcePoolID    uuid.UUID  `json:"resource_pool_id"`
	ResourcePoolName  string     `json:"resource_pool_name"`
	ProjectID         uuid.UUID  `json:"project_id"` // project id
	ProjectName       string     `json:"project_name"`
	NodeID            uuid.UUID  `json:"node_id"`
	NodeName          string     `json:"node_name"`
	GlideletURN       string     `json:"glidelet_urn"` // resource_pool.URN (where to send request)
	GlideletName      string     `json:"glidelet_name"`
	OrganizationName  string     `json:"organization_name"`
	Spec              GliderSpec `json:"spec"`
	ReferenceTicketID uuid.UUID  `json:"reference_ticket_id"`
	RedeemTimeout     uint       `json:"redeem_timeout"` // in seconds
	Lease             uint       `json:"lease"`          // in seconds
	CreatedAt         time.Time  `json:"created_at"`
}

// Value implements driver.Valuer for JSONB serialization
func (g GliderTicket) Value() (driver.Value, error) {
	return json.Marshal(g)
}

// Scan implements sql.Scanner for JSONB deserialization
func (g *GliderTicket) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	return json.Unmarshal(bytes, g)
}

type GliderSpec struct {
	Type      ResourceUnitType `json:"type"`
	PoolID    string           `json:"pool_id"`
	Resources SpecResourceList `json:"resource" gorm:"type:jsonb"`
}

type SpecResource struct {
	ResourceID string `json:"resource_id"`
	Name       string `json:"name"`
	Quantity   uint   `json:"quantity"`
	Unit       string `json:"unit"`
}
type SpecResourceList []SpecResource

func (s SpecResourceList) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *SpecResourceList) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSON value: %v", value)
	}
	return json.Unmarshal(bytes, s)
}

type StatusTicket string

const (
	StatusReady     StatusTicket = "ready"
	StatusRedeemed  StatusTicket = "redeemed"
	StatusPending   StatusTicket = "pending"
	StatusCancelled StatusTicket = "cancelled"
	StatusStopped   StatusTicket = "stopped"
	StatusExpired   StatusTicket = "expired"
	StatusFailed    StatusTicket = "failed"
)

var UneditableStatus = []StatusTicket{StatusCancelled, StatusStopped, StatusExpired, StatusFailed}

type ResourceUnitType string

const (
	ResourceUnitTypeCPU ResourceUnitType = "compute"
)
