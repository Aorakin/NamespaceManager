package usecase

import (
	"fmt"

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

func (u *NSUsecase) Update(editNS dtos.ReqForEdit) error {
	//ยังไม่ได้เช็คrole สำหรับการทำการupdate
	namespace := &dtos.EditNS{
		Priority: editNS.Priority,
		Quota:    editNS.Quota,
	}

	return u.NsRepository.Update(editNS.ID, namespace, editNS.UserIDs)
}

func (u *NSUsecase) Delete(namespaceID uuid.UUID) error {
	return u.NsRepository.Delete(namespaceID)
}

func (s *NSUsecase) AddUsersToNamespace(req dtos.UpdateNamespaceUsersReq) error {
	if len(req.UserIDs) == 0 {
		return fmt.Errorf("at least one user ID must be provided")
	}
	return s.NsRepository.AddUsersToNamespace(req.NamespaceID, req.UserIDs)
}

func (s *NSUsecase) RemoveUsersFromNamespace(req dtos.UpdateNamespaceUsersReq) error {
	fmt.Println(req)
	if len(req.UserIDs) == 0 {
		return fmt.Errorf("at least one user ID must be provided")
	}
	return s.NsRepository.RemoveUsersFromNamespace(req.NamespaceID, req.UserIDs)
}
