package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type GliderTicket struct {
	ID                uuid.UUID  `json:"id"`
	NamespaceURN      string     `json:"namespace_urn"` // namespace.id
	NamespaceName     string     `json:"namespace_name"`
	ProjectURN        string     `json:"project_urn"` // project.URN
	ProjectName       string     `json:"project_name"`
	GlideletURN       string     `json:"glidelet_urn"` // resource pool URN
	GlideletName      string     `json:"glidelet_name"`
	OrganizationName  string     `json:"organization_name"`
	Spec              GliderSpec `json:"spec" gorm:"type:jsonb"`
	ReferenceTicketID string     `json:"reference_ticket_id"`
	RedeemTimeout     uint       `json:"redeem_timeout"` // in seconds
	Lease             uint       `json:"lease"`          // in seconds
	CreatedAt         time.Time  `json:"created_at"`
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
