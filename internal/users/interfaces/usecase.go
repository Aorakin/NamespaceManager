package interfaces

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/users/dtos"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UsersUsecase interface {
	GenerateLoginURL(string) string
	HandleGoogleCallback(string, *gin.Context) (map[string]interface{}, error)
	Register(dtos.RegisterInput) error
	Login(string, string) (*models.User, error)
	GetUserByID(string) (*models.User, error)
	FindOrCreateUser(uuid.UUID, string, string, string) (*models.User, error)
}
