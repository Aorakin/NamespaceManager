package usecase

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/namespace/dtos"
	"github.com/NamespaceManager/internal/namespace/interfaces"
)

type NSUsecase struct {
	NsRepository interfaces.NSRepository
}

func NewNSUsecase(nsRepository interfaces.NSRepository) interfaces.NSUsecase {
	return &NSUsecase{
		NsRepository: nsRepository,
	}
}

func (u *NSUsecase) HandleCreate(request dtos.RequestNS) error {
	namespace := &models.Namespace{
		URN:              request.URN,
		ProjectURN:       request.ProjectURN,
		Priority:         request.Priority,
		Quota:            request.Quota,
		ResourceUnitURNs: request.ResourceUnitURNs,
	}
	return u.NsRepository.Create(*namespace)
}
