package dtos

type RequestTicket struct {
	GPU string `json:"gpu"`
	Ram string `json:"ram"`
	CPU string `json:"cpu"`
}
