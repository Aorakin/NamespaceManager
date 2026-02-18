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
	startFailedTicketIDs := make([]uuid.UUID, 0)

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

		if hasStartFailed {
			// Mark this ticket for special handling
			startFailedTicketIDs = append(startFailedTicketIDs, statusRes.TicketID)
			// Don't add to ticketUpdates - handleStartFailedTickets will handle the update
			continue
		}

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

	// Handle start_failed tickets: create dummy tickets and reassign to tasks
	// Returns the task IDs that were affected
	affectedTaskIDs := make(map[uuid.UUID]struct{})
	if len(startFailedTicketIDs) > 0 {
		taskIDs, err := u.handleStartFailedTickets(startFailedTicketIDs)
		if err != nil {
			return apiError.NewInternalServerError(fmt.Errorf("failed to handle start_failed tickets: %w", err))
		}
		for _, taskID := range taskIDs {
			affectedTaskIDs[taskID] = struct{}{}
		}
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

	// Add the task IDs affected by start_failed tickets
	for taskID := range affectedTaskIDs {
		taskIDsMap[taskID] = struct{}{}
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
	tickets, err := u.ticketRepository.GetTicketsByTaskID(taskID, true) // include failed tickets to determine if task should be marked as failed
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

// handleStartFailedTickets creates dummy tickets for failed starts and manages task assignments
// Returns the list of task IDs that were affected
func (u *TicketUsecase) handleStartFailedTickets(ticketIDs []uuid.UUID) ([]uuid.UUID, error) {
	// Get original tickets
	originalTickets, err := u.ticketRepository.GetTicketsByGliderTicketIDs(ticketIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get original tickets: %w", err)
	}

	affectedTaskIDs := make([]uuid.UUID, 0)

	for _, originalTicket := range originalTickets {
		if originalTicket.TaskID == nil {
			continue
		}

		taskID := *originalTicket.TaskID
		affectedTaskIDs = append(affectedTaskIDs, taskID)

		// Create a dummy ticket
		dummyGliderTicketID := uuid.New()

		// Copy GliderTicket and update its ID
		dummyGliderTicket := originalTicket.GliderTicket
		dummyGliderTicket.ID = dummyGliderTicketID

		dummyTicket := models.Ticket{
			Name:           originalTicket.Name,
			GliderTicket:   dummyGliderTicket,
			GliderTicketID: dummyGliderTicketID, // New unique ID for dummy ticket
			Signature:      originalTicket.Signature,
			Status:         models.StatusFailed,
			TaskID:         &taskID, // Keep in the same task
			OwnerID:        originalTicket.OwnerID,
			OwnerName:      originalTicket.OwnerName,
			NamespaceID:    originalTicket.NamespaceID,
			GlideletURN:    originalTicket.GlideletURN,
			ResourcePoolID: originalTicket.ResourcePoolID,
			URL:            originalTicket.URL,
			Password:       originalTicket.Password,
			Failed:         true,
		}

		// Create the dummy ticket in database
		if err := u.ticketRepository.Create(&dummyTicket); err != nil {
			return nil, fmt.Errorf("failed to create dummy ticket for %s: %w", originalTicket.GliderTicketID, err)
		}

		// Update original ticket: reset to ready and clear task assignment
		originalTicket.Status = models.StatusReady
		originalTicket.TaskID = nil
		if err := u.ticketRepository.Update(originalTicket); err != nil {
			return nil, fmt.Errorf("failed to update original ticket %s: %w", originalTicket.GliderTicketID, err)
		}
	}

	return affectedTaskIDs, nil
}
