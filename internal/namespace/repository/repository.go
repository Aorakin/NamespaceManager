package repository

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/namespace/interfaces"
	"gorm.io/gorm"
)

type NSRepository struct {
	db *gorm.DB
}

func NewUsersRepository(db *gorm.DB) interfaces.NSRepository {
	return &NSRepository{db: db}
}

func (r *NSRepository) Create(namespace models.Namespace) error {
	if err := r.db.Create(&namespace).Error; err != nil {
		return err
	}
	return nil
}
