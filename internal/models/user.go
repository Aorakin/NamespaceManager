package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;unique" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (base *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	return
}

type ProviderType string

const (
	ProviderLocal  ProviderType = "local"
	ProviderGoogle ProviderType = "google"
	ProviderGithub ProviderType = "github"
)

type User struct {
	BaseModel
	Username      string         `gorm:"not null;uniqueIndex" json:"username" validate:"required,min=2,max=30"`
	Email         string         `gorm:"not null;unique" json:"email" validate:"required,email"`
	Password      string         `gorm:"" json:"password,omitempty" validate:"omitempty,min=8"` // Made optional for OAuth users
	Role          UserRole       `gorm:"type:varchar(50);default:'user'" json:"role"`
	// Namespaces    []*Namespace   `gorm:"many2many:user_namespaces;" json:"namespace"`
	UserProviders []UserProvider `gorm:"foreignKey:UserID" json:"providers,omitempty"`
}

type UserProvider struct {
	BaseModel
	UserID       uuid.UUID    `gorm:"type:uuid;not null" json:"user_id"`
	User         User         `gorm:"foreignKey:UserID" json:"-"`
	Provider     ProviderType `gorm:"type:varchar(50);not null" json:"provider"`
	ProviderID   string       `gorm:"not null" json:"provider_id"`           // The ID from the OAuth provider (Google ID, etc.)
	AccessToken  string       `gorm:"type:text" json:"-"`                    // Store access token (hidden from JSON)
	RefreshToken string       `gorm:"type:text" json:"-"`                    // Store refresh token (hidden from JSON)
	ExpiresAt    *time.Time   `json:"expires_at,omitempty"`                  // Token expiration time
}

// Helper methods for OAuth integration
func (u *User) HasProvider(provider ProviderType) bool {
	for _, up := range u.UserProviders {
		if up.Provider == provider {
			return true
		}
	}
	return false
}

func (u *User) GetProvider(provider ProviderType) *UserProvider {
	for _, up := range u.UserProviders {
		if up.Provider == provider {
			return &up
		}
	}
	return nil
}

func (u *User) AddProvider(provider ProviderType, providerID string, accessToken, refreshToken string, expiresAt *time.Time) *UserProvider {
	userProvider := UserProvider{
		UserID:       u.ID,
		Provider:     provider,
		ProviderID:   providerID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}
	u.UserProviders = append(u.UserProviders, userProvider)
	return &userProvider
}
