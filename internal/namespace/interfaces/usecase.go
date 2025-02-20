package interfaces

import (
	"github.com/NamespaceManager/internal/namespace/dtos"
	"github.com/google/uuid"
)

type NSUsecase interface {
	HandleCreate(dtos.RequestNS, uuid.UUID) error
	GetNSList(uuid.UUID) ([]dtos.NSresponse, error)
}
