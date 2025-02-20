package interfaces

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/google/uuid"
)

type NSRepository interface {
	Create(models.Namespace, uuid.UUID) error
	GetNsList(uuid.UUID) ([]*models.Namespace, error)
}
