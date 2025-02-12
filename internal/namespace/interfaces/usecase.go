package interfaces

import "github.com/NamespaceManager/internal/namespace/dtos"

type NSUsecase interface {
	HandleCreate(dtos.RequestNS) error
}
