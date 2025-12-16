package repository

import (
	"errors"

	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/namespace/dtos"
	"github.com/NamespaceManager/internal/namespace/interfaces"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NamespaceRepository struct {
	db *gorm.DB
}

func NewNamespaceRepository(db *gorm.DB) interfaces.NamespaceRepository {
	return &NamespaceRepository{db: db}
}

func (r *NamespaceRepository) Create(namespace models.Namespace, userID uuid.UUID) error {
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

func (r *NamespaceRepository) GetNsList(userID uuid.UUID) ([]*models.Namespace, error) {
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

func (r *NamespaceRepository) Delete(namespaceID uuid.UUID) error {
	var namespace models.Namespace
	if err := r.db.First(&namespace, namespaceID).Error; err != nil {
		return err
	}

	if err := r.db.Model(&namespace).Association("Users").Clear(); err != nil {
		return err
	}

	if err := r.db.Delete(&namespace).Error; err != nil {
		return err
	}

	return nil
}

func (r *NamespaceRepository) Update(namespaceID uuid.UUID, updatedData *dtos.EditNS, userIDs []uuid.UUID) error {
	var namespace models.Namespace

	if err := r.db.Model(&namespace).Updates(updatedData).Error; err != nil {
		return err
	}
	namespace, users, err := r.FetchNamespace(namespaceID, userIDs)
	if err != nil {
		return err
	}
	if err := r.db.Model(&namespace).Association("Users").Replace(users); err != nil {
		return err
	}

	return nil
}

func (r *NamespaceRepository) RemoveUsersFromNamespace(namespaceID uuid.UUID, userIDs []uuid.UUID) error {
	namespace, users, err := r.FetchNamespace(namespaceID, userIDs)
	if err != nil {
		return err
	}

	if err := r.db.Model(&namespace).Association("Users").Delete(users); err != nil {
		return err
	}

	return nil
}

func (r *NamespaceRepository) AddUsersToNamespace(namespaceID uuid.UUID, userIDs []uuid.UUID) error {
	namespace, users, err := r.FetchNamespace(namespaceID, userIDs)
	if err != nil {
		return err
	}

	if err := r.db.Model(&namespace).Association("Users").Append(users); err != nil {
		return err
	}

	return nil
}

func (r *NamespaceRepository) FetchNamespace(namespaceID uuid.UUID, userIDs []uuid.UUID) (models.Namespace, []models.User, error) {
	var namespace models.Namespace

	if err := r.db.Preload("Users").First(&namespace, namespaceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return namespace, nil, errors.New("namespace not found")
		}
		return namespace, nil, err
	}

	var users []models.User
	if err := r.db.Where("id IN (?)", userIDs).Find(&users).Error; err != nil {
		return namespace, nil, err
	}

	return namespace, users, nil
}
