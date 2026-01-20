package usecase

import (
	"fmt"
	"strings"

	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/ticket/dtos"
	apiError "github.com/NamespaceManager/pkg/api_error"
	"github.com/google/uuid"
)

func (u *TicketUsecase) UpdateTicketStatusFromGlidelet(req []dtos.StatusRes) error {
	if len(req) == 0 {
		return nil
	}

	ticketCodeServerUpdates := dtos.CodeServerResponse{TicketResponse: []dtos.TicketResponse{}}
	ticketUpdates := make(map[uuid.UUID]models.StatusTicket)
	ticketIDs := make([]uuid.UUID, 0, len(req))

	for _, statusRes := range req {
		ticketIDs = append(ticketIDs, statusRes.TicketID)

		if statusRes.HasError {
			ticketUpdates[statusRes.TicketID] = models.StatusFailed
			continue
		}

		if len(statusRes.PodStatus) == 0 {
			fmt.Printf("Warning: No pod status found for ticketId: %s. Skipping update.", statusRes.TicketID)
			continue
		}

		if statusRes.CodeServerURL != "" {
			ticketCodeServerUpdates.TicketResponse = append(ticketCodeServerUpdates.TicketResponse, dtos.TicketResponse{
				TicketID: statusRes.TicketID,
				URL:      statusRes.CodeServerURL,
				Password: statusRes.Password,
			})
		}
		finalStatus := u.computeTicketStatus(statusRes.PodStatus)
		ticketUpdates[statusRes.TicketID] = finalStatus
	}

	if err := u.ticketRepository.BatchUpdateTicketStatuses(ticketUpdates); err != nil {
		return apiError.NewInternalServerError(fmt.Errorf("failed to batch update ticket statuses: %w", err))
	}

	if err := u.updateTicketInfo(&ticketCodeServerUpdates); err != nil {
		return err
	}

	tickets, err := u.ticketRepository.GetTicketsByGliderTicketIDs(ticketIDs)
	if err != nil {
		return apiError.NewInternalServerError(fmt.Errorf("failed to retrieve tickets: %w", err))
	}

	taskIDsMap := make(map[uuid.UUID]struct{})
	for _, ticket := range tickets {
		if ticket.TaskID != nil {
			taskIDsMap[*ticket.TaskID] = struct{}{}
		}
	}

	for taskID := range taskIDsMap {
		if err := u.updateTaskStatus(taskID); err != nil {
			return apiError.NewInternalServerError(fmt.Errorf("failed to update task status for task %s: %v", taskID, err))
		}
	}

	nodeNamesSet := make(map[string]struct{})
	for _, ticket := range tickets {
		if ticket.Status == models.StatusExpired || ticket.Status == models.StatusFailed {
			nodeNamesSet[ticket.GliderTicket.NodeName] = struct{}{}
		}
	}

	nodeNames := make([]string, 0, len(nodeNamesSet))
	for nodeName := range nodeNamesSet {
		nodeNames = append(nodeNames, nodeName)
	}

	u.requeue(nodeNames)

	return nil
}

// computeTicketStatus determines the final ticket status based on pod statuses
func (u *TicketUsecase) computeTicketStatus(podStatuses []dtos.PodStatus) models.StatusTicket {
	ticketRunning := true

	for _, pod := range podStatuses {
		podStatus := strings.ToLower(pod.Status)

		if podStatus == "inactive" {
			return models.StatusExpired
		}

		if podStatus == "pending" {
			return models.StatusPending
		}

		if podStatus != "running" {
			ticketRunning = false
		}
	}

	if ticketRunning {
		return models.StatusRedeemed
	}

	return models.StatusFailed
}

func (u *TicketUsecase) updateTaskStatus(taskID uuid.UUID) error {
	tickets, err := u.ticketRepository.GetTicketsByTaskID(taskID)
	if err != nil {
		return apiError.NewInternalServerError(fmt.Errorf("failed to get tickets by task ID: %w", err))
	}

	if len(tickets) == 0 {
		return nil
	}

	var (
		anyFailed    bool
		anyStopped   bool
		anyPending   bool
		anyRedeemed  bool
		expiredCount int
	)

	for _, t := range tickets {
		switch t.Status {
		case models.StatusFailed:
			anyFailed = true

		case models.StatusStopped:
			anyStopped = true

		case models.StatusPending:
			anyPending = true

		case models.StatusExpired:
			expiredCount++

		case models.StatusRedeemed:
			anyRedeemed = true
		}
	}

	allExpired := expiredCount == len(tickets)

	var taskStatus models.StatusTicket

	// Apply priority: any failed > any stopped > any pending > any redeemed (running) > all expired
	switch {
	case anyFailed:
		taskStatus = models.StatusFailed
	case anyStopped:
		taskStatus = models.StatusStopped
	case anyPending:
		taskStatus = models.StatusPending
	case anyRedeemed:
		taskStatus = models.StatusRedeemed
	case allExpired:
		taskStatus = models.StatusExpired
	default:
		taskStatus = models.StatusFailed
	}

	return u.ticketRepository.UpdateTaskStatus(taskID, taskStatus)
}
