package dtos

// ProjectDTO represents a project with its details
type ProjectDTO struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	OrganizationID string `json:"organization_id"`
}

// NamespaceDTO represents a namespace with its details
type NamespaceDTO struct {
	ID               string             `json:"id"`
	Name             string             `json:"name"`
	Description      string             `json:"description"`
	Credit           float64            `json:"credit"`
	ProjectID        string             `json:"project_id"`
	Quotas           string             `json:"quotas"`
	NamespaceMembers []NamespaceMembers `json:"namespace_members"`
}
type NamespaceMembers struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// QuotaDTO represents quota details for a namespace
type QuotaDTO struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	ResourcePoolID   string          `json:"resource_pool_id"`
	ResourcePoolName string          `json:"resource_pool_name"`
	OrganizationName string          `json:"organization_name"`
	ProjectID        string          `json:"project_id"`
	Resources        []QuotaResource `json:"resources"`
}
type QuotaResource struct {
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

// UsageDTO represents usage details for a project or namespace
type UsageDTO struct {
	Usage []UsageDetail `json:"usage"`
}
type UsageDetail struct {
	TypeID string `json:"type_id"`
	Type   string `json:"type"`
	Quota  int64  `json:"quota"`
	Usage  int64  `json:"usage"`
}

// ResourceDTO represents a resource with its details
type ResourceDTO struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Quantity       int64        `json:"quantity"`
	ResourceTypeID string       `json:"resource_type_id"`
	ResourceType   ResourceType `json:"resource_type"`
	ResourcePoolID string       `json:"resource_pool_id"`
}
type ResourceType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Unit string `json:"unit"`
}

type ResourcesPoolDetailDTO struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	DisplayName    string `json:"display_name"`
	OrganizationID string `json:"organization_id"`
}
