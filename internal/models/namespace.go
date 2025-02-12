package models

type Namespace struct {
	BaseModel
	URN              string         `gorm:"uniqueIndex;not null" json:"urn"`
	ProjectURN       string         `gorm:"not null" json:"project"`
	Users            []*User        `gorm:"many2many:user_namespaces;"`
	Priority         string         `gorm:"not null" json:"priority"`
	Quota            string         `gorm:"not null" json:"quota"`
	ResourceUnitURNs string         `gorm:"not null" json:"resource_unit_urn"`
	Tickets          []GliderTicket `gorm:"foreignKey:NamespaceID" json:"spec" validate:"required,min"`
}
