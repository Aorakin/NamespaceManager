package dtos

type ProjectDTO struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	OrganizationID string `json:"organization_id"`
}
type NamespaceDTO struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Credit           float64  `json:"credit"`
	ProjectID        string   `json:"project_id"`
	Quotas           string   `json:"quotas"`
	NamespaceMembers []NamespaceMembers `json:"namespace_members"`
}
type NamespaceMembers struct {
	ID					string `json:"id"`
	Email				string `json:"email"`
	FirstName			string `json:"first_name"`
	LastName			string `json:"last_name"`
}
type QuotaDTO struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	ProjectID      string     `json:"project_id"`
	ProjectQuotaID string     `json:"project_quota_id"`
	ResourcePoolID string     `json:"resource_pool_id"`
	Resources      []Resource `json:"resources"`
}
type Resource struct {
	ID                 string           `json:"id"`
	NamespaceQuotaID   string           `json:"namespace_quota_id"`
	Quantity           int64            `json:"quantity"`
	ResourcePropertyID string           `json:"resource_property_id"`
	ResourceProperty   ResourceProperty `json:"resource_prop"`
}
type ResourceProperty struct {
	ID          string `json:"id"`
	ResourceID  string `json:"resource_id"`
	Price       int64  `json:"price"`
	MaxDuration int64  `json:"max_duration"`
}
type UsageDTO struct {
	Usage []UsageDetail `json:"usage"`
}
type UsageDetail struct {
	TypeID string `json:"type_id"`
	Type   string `json:"type"`
	Quota  int64  `json:"quota"`
	Usage  int64  `json:"usage"`
}
