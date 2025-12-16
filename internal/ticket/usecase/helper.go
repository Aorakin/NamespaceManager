package usecase

// func (u *TicketUsecase) isNamespaceMember(namespaceID uuid.UUID, userID uuid.UUID) error {
// 	namespace, err := u.namespaceRepo.GetByID(namespaceID)
// 	if err != nil {
// 		return apiError.NewNotFoundError(fmt.Errorf("namespace not found: %w", err))
// 	}

// 	user, err := u.userRepo.GetUser(userID)
// 	if err != nil {
// 		return apiError.NewNotFoundError(fmt.Errorf("user not found: %w", err))
// 	}

// 	if !helper.ContainsUserID(namespace.Members, user.ID) {
// 		return apiError.NewForbiddenError(fmt.Errorf("user is not a member of the namespace"))
// 	}

// 	return nil
// }
