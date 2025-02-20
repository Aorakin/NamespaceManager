package repository

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/namespace/interfaces"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NSRepository struct {
	db *gorm.DB
}

func NewUsersRepository(db *gorm.DB) interfaces.NSRepository {
	return &NSRepository{db: db}
}

func (r *NSRepository) Create(namespace models.Namespace, userID uuid.UUID) error {
	if err := r.db.Create(&namespace).Error; err != nil {
		return err
	}

	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return err
	}

	if err := r.db.Model(&namespace).Association("Users").Append(&user); err != nil {
		return err
	}

	return nil
}

func (r *NSRepository) GetNsList(userID uuid.UUID) ([]*models.Namespace, error) {
	var NSlist []*models.Namespace
	err := r.db.
		Joins("JOIN user_namespaces ON user_namespaces.namespace_id = namespaces.id").
		Where("user_namespaces.user_id = ?", userID).
		Find(&NSlist).
		Order("urn ASC").Error
	if err != nil {
		return nil, err
	}
	return NSlist, nil
}
