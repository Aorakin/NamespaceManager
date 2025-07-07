package dtos

type RequestTicket struct {
	PoolName  string `json:"pool_name" validate:"required" `
	GPU       string `json:"gpu" `
	Ram       string `json:"ram"`
	VRam      string `json:"vram"`
	CPU       string `json:"cpu"`
	Storage   string `json:"storage"`
	UsageTime string `json:"usage_time" validate:"required"`
}
