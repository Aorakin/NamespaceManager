package dtos

import "github.com/google/uuid"

type StatusRes struct {
	TicketID  uuid.UUID   `json:"ticketId"`
	PodStatus []PodStatus `json:"status"`
	HasError  bool        `json:"hasError"`
}

type PodStatus struct {
	PodID    string `json:"podId"`
	Status   string `json:"status"`
	ErrorMsg string `json:"errorMsg,omitempty"`
}
