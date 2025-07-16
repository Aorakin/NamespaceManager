package dtos

import "github.com/google/uuid"

type NSresponse struct {
	ID               uuid.UUID `json:"id"`
	URN              string    `gorm:"uniqueIndex;not null" json:"urn"`
	ProjectURN       string    `gorm:"not null" json:"project"`
	Priority         string    `gorm:"not null" json:"priority"`
	Quota            string    `gorm:"not null" json:"quota"`
	ResourceUnitURNs string    `gorm:"not null" json:"resource_unit_urn"`
	// Spec             string    `json:"spec"`
}
