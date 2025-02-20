package usecase

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/namespace/dtos"
	"github.com/NamespaceManager/internal/namespace/interfaces"
	"github.com/google/uuid"
)

type NSUsecase struct {
	NsRepository interfaces.NSRepository
}

func NewNSUsecase(nsRepository interfaces.NSRepository) interfaces.NSUsecase {
	return &NSUsecase{
		NsRepository: nsRepository,
	}
}

func (u *NSUsecase) HandleCreate(request dtos.RequestNS, userID uuid.UUID) error {
	namespace := &models.Namespace{
		URN:              request.URN,
		ProjectURN:       request.ProjectURN,
		Priority:         request.Priority,
		Quota:            request.Quota,
		ResourceUnitURNs: request.ResourceUnitURNs,
	}

	return u.NsRepository.Create(*namespace, userID)
}

func (u *NSUsecase) GetNSList(userID uuid.UUID) ([]dtos.NSresponse, error) {
	var resp []dtos.NSresponse
	NSlist, err := u.NsRepository.GetNsList(userID)
	if err != nil {
		return nil, err
	}

	if len(NSlist) != 0 {
		resp = make([]dtos.NSresponse, len(NSlist))
		for i, ns := range NSlist {
			resp[i] = dtos.NSresponse{
				ID:               ns.ID,
				URN:              ns.URN,
				ProjectURN:       ns.ProjectURN,
				Quota:            ns.Quota,
				ResourceUnitURNs: ns.ResourceUnitURNs,
				Priority:         ns.Priority,
			}
		}
	}

	return resp, nil
}
