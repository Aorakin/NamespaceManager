package usecase

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/ticket/dtos"
	apiError "github.com/NamespaceManager/pkg/api_error"
	"github.com/NamespaceManager/pkg/httpclient"
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

		// Check if any pod has "start_failed" status first (before hasError check)
		hasStartFailed := false
		if len(statusRes.PodStatus) > 0 {
			for _, pod := range statusRes.PodStatus {
				if strings.ToLower(pod.Status) == "start_failed" {
					hasStartFailed = true
					break
				}
			}
		}

		// if hasStartFailed {
		// 	// Mark this ticket for special handling
		// 	startFailedTicketIDs = append(startFailedTicketIDs, statusRes.TicketID)
		// 	// Don't add to ticketUpdates - handleStartFailedTickets will handle the update
		// 	continue
		// }

		if statusRes.HasError || hasStartFailed {
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

	// // Handle start_failed tickets: create dummy tickets and reassign to tasks
	// // Returns the task IDs that were affected
	// affectedTaskIDs := make(map[uuid.UUID]struct{})
	// if len(startFailedTicketIDs) > 0 {
	// 	taskIDs, err := u.handleStartFailedTickets(startFailedTicketIDs)
	// 	if err != nil {
	// 		return apiError.NewInternalServerError(fmt.Errorf("failed to handle start_failed tickets: %w", err))
	// 	}
	// 	for _, taskID := range taskIDs {
	// 		affectedTaskIDs[taskID] = struct{}{}
	// 	}
	// }

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

	// Get Task IDs affected by status changes
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
	tickets, err := u.ticketRepository.GetTicketsByTaskID(taskID, true)
	if err != nil {
		return apiError.NewInternalServerError(fmt.Errorf("failed to get tickets by task ID: %w", err))
	}

	if len(tickets) == 0 {
		return nil
	}

	task, err := u.ticketRepository.GetTasksByID(taskID)
	if err != nil {
		return apiError.NewInternalServerError(fmt.Errorf("failed to get task: %w", err))
	}

	var (
		anyFailed   bool
		anyStopped  bool
		anyPending  bool
		anyRedeemed bool
		anyExpired  bool
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
			anyExpired = true
		case models.StatusRedeemed:
			anyRedeemed = true
		}
	}

	allRedeemed := anyRedeemed && !anyFailed && !anyStopped && !anyPending && !anyExpired
	// update task start time
	if allRedeemed && task.StartedAt == nil {
		if err := u.ticketRepository.UpdateTaskStartTime(taskID, time.Now()); err != nil {
			return apiError.NewInternalServerError(fmt.Errorf("failed to update task start time: %w", err))
		}
	}

	alreadyActivated := task.StartedAt != nil

	var taskStatus models.StatusTicket

	if !alreadyActivated {
		switch {
		case allRedeemed:
			taskStatus = models.StatusRedeemed
		case anyPending && !anyFailed && !anyStopped && !anyExpired:
			taskStatus = models.StatusPending
		default:
			if err := u.clusterRollback(taskID); err != nil {
				log.Printf("[UPDATE TASK STATUS] Cluster rollback failed for task %s: %v", taskID, err)
				return apiError.NewInternalServerError(fmt.Errorf("failed to rollback cluster: %w", err))
			}
			taskStatus = models.StatusFailed
		}
	} else {
		switch {
		case anyRedeemed:
			taskStatus = models.StatusRedeemed
		case anyFailed:
			taskStatus = models.StatusFailed
		case anyStopped:
			taskStatus = models.StatusStopped
		case anyExpired:
			taskStatus = models.StatusExpired
		default:
			taskStatus = models.StatusFailed
		}
	}

	return u.ticketRepository.UpdateTaskStatus(taskID, taskStatus)
}
func (u *TicketUsecase) clusterRollback(taskID uuid.UUID) error {
	tickets, err := u.ticketRepository.GetTicketsByTaskID(taskID, true)
	if err != nil {
		return apiError.NewInternalServerError(fmt.Errorf("failed to get tickets for cluster rollback: %w", err))
	}

	// 1. Rollback actual cluster resources
	u.rollbackClusterResources(tickets)

	// 2. Create dummy failed tickets and reset originals
	u.createDummyTicketsAndReset(taskID, tickets)

	return nil
}
func (u *TicketUsecase) createDummyTicketsAndReset(taskID uuid.UUID, tickets []models.Ticket) error {
	ticketUpdates := make(map[uuid.UUID]models.StatusTicket)
	ticketIDs := make([]uuid.UUID, 0, len(tickets))

	for _, ticket := range tickets {
		dummyGliderTicketID := uuid.New()

		dummyGliderTicket := ticket.GliderTicket
		dummyGliderTicket.ID = dummyGliderTicketID

		dummyTicket := models.Ticket{
			Name:           ticket.Name,
			GliderTicket:   dummyGliderTicket,
			GliderTicketID: dummyGliderTicketID,
			Signature:      ticket.Signature,
			Status:         models.StatusFailed,
			TaskID:         &taskID,
			OwnerID:        ticket.OwnerID,
			OwnerName:      ticket.OwnerName,
			NamespaceID:    ticket.NamespaceID,
			GlideletURN:    ticket.GlideletURN,
			ResourcePoolID: ticket.ResourcePoolID,
			URL:            ticket.URL,
			Password:       ticket.Password,
			Failed:         true,
		}

		if err := u.ticketRepository.Create(&dummyTicket); err != nil {
			log.Printf("[DUMMY TICKET] Failed to create dummy ticket for %s: %v", ticket.GliderTicketID, err)
			continue
		}

		ticketUpdates[ticket.ID] = models.StatusReady
		ticketIDs = append(ticketIDs, ticket.ID)
	}

	if err := u.ticketRepository.BatchUpdateTicketStatuses(ticketUpdates); err != nil {
		log.Printf("[DUMMY TICKET] Failed to batch update ticket statuses: %v", err)
	}

	if err := u.ticketRepository.BatchClearTaskIDs(ticketIDs); err != nil {
		log.Printf("[DUMMY TICKET] Failed to batch clear task IDs: %v", err)
	}

	return nil
}

func (u *TicketUsecase) rollbackClusterResources(tickets []models.Ticket) error {
	ticketIDs := []uuid.UUID{}
	for _, ticket := range tickets {
		ticketIDs = append(ticketIDs, ticket.GliderTicketID)
	}

	ticketsByPool, err := u.groupTicketByPool(tickets)
	if err != nil {
		return apiError.NewInternalServerError(fmt.Errorf("failed to group ticket IDs by pool: %w", err))
	}

	for poolID, poolTickets := range ticketsByPool {
		poolURN, err := u.getPoolURN(poolID, poolTickets[0].GlideletURN)
		if err != nil {
			log.Printf("[ROLLBACK] Failed to get pool URN for pool %s: %v", poolID, err)
			continue
		}

		payload := dtos.ResetTicketRequest{TicketIDs: ticketIDs}
		_, err = httpclient.SendRequest(poolURN+"/api/v1/ticket/rollbackPods", payload, "PATCH")
		if err != nil {
			log.Printf("[ROLLBACK] Failed to send rollback request to pool %s: %v", poolID, err)
			continue
		}

		log.Printf("[ROLLBACK] Successfully rolled back pool %s", poolID)
	}

	return nil
}
