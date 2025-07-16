package interfaces

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/namespace/dtos"
	"github.com/google/uuid"
)

type NSRepository interface {
	Create(models.Namespace, uuid.UUID) error
	GetNsList(uuid.UUID) ([]*models.Namespace, error)
	Update(uuid.UUID, *dtos.EditNS, []uuid.UUID) error
	Delete(uuid.UUID) error
	AddUsersToNamespace(namespaceID uuid.UUID, userIDs []uuid.UUID) error
	RemoveUsersFromNamespace(namespaceID uuid.UUID, userIDs []uuid.UUID) error
}
