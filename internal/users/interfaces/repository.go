package interfaces

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type UsersRepository interface {
	GetUserGoogle(*oauth2.Token) (map[string]interface{}, error)
	Create(*models.User) error
	Delete(uuid.UUID) error
	GetUser(uuid.UUID) (*models.User, error)
	GetByUsername(string) (*models.User, error)
	GetByEmail(string) (*models.User, error)
	CreateOrGetOAuthUser(providerID, email, username string, provider models.ProviderType, token *oauth2.Token) (*models.User, error)

	FindOrCreateUser(uuid.UUID, string, string, string) (*models.User, error)
}
