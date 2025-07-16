package dtos

import "github.com/google/uuid"

type EditNS struct {
	Priority string `gorm:"not null" json:"priority"`
	Quota    string `gorm:"not null" json:"quota"`
}

type ReqForEdit struct {
	ID       uuid.UUID   `gorm:"not null" json:"id"`
	Priority string      `gorm:"not null" json:"priority"`
	Quota    string      `gorm:"not null" json:"quota"`
	UserIDs  []uuid.UUID `gorm:"not null" json:"user_ids"`
}

type UpdateNamespaceUsersReq struct {
	UserIDs     []uuid.UUID `json:"user_ids" validate:"required"`
	NamespaceID uuid.UUID   `json:"namespace_id" validate:"required"`
}
