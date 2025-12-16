package repository

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/google/uuid"
)

func (r *NamespaceRepository) GetByID(namespaceID uuid.UUID) (*models.Namespace, error) {
	var namespace models.Namespace
	if err := r.db.Preload("Users").First(&namespace, "id = ?", namespaceID).Error; err != nil {
		return nil, err
	}
	return &namespace, nil
}
