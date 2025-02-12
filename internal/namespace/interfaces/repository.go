package interfaces

import "github.com/NamespaceManager/internal/models"

type NSRepository interface {
	Create(models.Namespace) error
}
