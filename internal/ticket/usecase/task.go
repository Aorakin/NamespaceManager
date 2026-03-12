package usecase

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/ticket/dtos"
	"github.com/NamespaceManager/internal/ticket/mapper"
	apiError "github.com/NamespaceManager/pkg/api_error"
	"github.com/NamespaceManager/pkg/httpclient"
	"github.com/google/uuid"
)

func (u *TicketUsecase) GetTasks(ownerID uuid.UUID) ([]models.Task, error) {
	tasks, err := u.ticketRepository.GetTasks(ownerID)
	if err != nil {
		log.Printf("failed to get tasks: %v", err)
		return nil, apiError.NewInternalServerError("Failed to retrieve tasks")
	}

	return tasks, nil
}

func (u *TicketUsecase) CreateTask(request *dtos.CreateTaskRequest, userID uuid.UUID) error {
	tickets, err := u.ticketRepository.GetTicketsByGliderTicketIDs(request.Tickets)
	if err != nil {
		log.Printf("failed to retrieve tickets: %v", err)
		return apiError.NewInternalServerError("Failed to retrieve tickets")
	}

	namespaceID := uuid.Nil

	for _, ticket := range tickets {
		if namespaceID == uuid.Nil {
			namespaceID = ticket.NamespaceID
		} else if ticket.NamespaceID != namespaceID {
			return apiError.NewBadRequestError("All tickets must belong to the same namespace")
		}

		if ticket.Status != models.StatusReady {
			return apiError.NewBadRequestError("One or more tickets are not in 'Ready' status")
		}
		if ticket.OwnerID != userID {
			return apiError.NewForbiddenError("You do not have permission to use one or more tickets")
		}
	}

	ticketsByPool, err := u.groupTicketIDsByPool(request.Tickets)
	if err != nil {
		return err
	}

	startTime, err := u.EnqueueTask(ticketsByPool)
	log.Printf("[CREATE TASK] EnqueueTask returned startTime: %v", startTime)
	if err != nil {
		log.Printf("[CREATE TASK] EnqueueTask returned with error: %s", err.Error())
		return err
	}
	status := models.StatusQueued

	if !startTime.IsZero() {
		status = models.StatusPending
		// If start time is more than 1 minute in the future, it's queued
		if time.Until(startTime) > time.Minute {
			status = models.StatusQueued
		}

		log.Printf("[CREATE TASK] Start time is scheduled at %v, status: %s", startTime, status)
		err = u.confirmTickets(ticketsByPool)
		if err != nil {
			return err
		}

	} else {
		log.Printf("[CREATE TASK] Start time is ZERO")
	}

	task := models.Task{
		Title:              request.Title,
		Tickets:            tickets,
		OwnerID:            userID,
		Status:             status,
		EstimatedStartTime: startTime,
	}

	return u.ticketRepository.CreateTask(task)
}

func (u *TicketUsecase) CancelTask(userID uuid.UUID, taskID uuid.UUID) error {
	task, err := u.ticketRepository.GetTasksByID(taskID)
	if err != nil {
		log.Printf("failed to get task by ID %s: %v", taskID, err)
		return apiError.NewInternalServerError("Failed to find the task")
	}
	if task.OwnerID != userID {
		return apiError.NewForbiddenError("You do not have permission to cancel this task")
	}

	// Group tickets by resource pool
	type poolInfo struct {
		ticketIDs  []uuid.UUID
		defaultURL string
	}
	ticketsByPool := make(map[uuid.UUID]*poolInfo)

	for _, ticket := range task.Tickets {
		poolID := ticket.ResourcePoolID
		if _, exists := ticketsByPool[poolID]; !exists {
			ticketsByPool[poolID] = &poolInfo{
				ticketIDs:  []uuid.UUID{},
				defaultURL: ticket.GlideletURN,
			}
		}
		ticketsByPool[poolID].ticketIDs = append(ticketsByPool[poolID].ticketIDs, ticket.GliderTicketID)
	}

	statusUpdates := make(map[uuid.UUID]models.StatusTicket)
	var cancelErrors []error

	// First, attempt to cancel all tickets in all pools
	for poolID, info := range ticketsByPool {
		poolURL, err := u.getPoolURN(poolID.String(), info.defaultURL)
		if err != nil {
			cancelErrors = append(cancelErrors, fmt.Errorf("pool %s URN: %w", poolID, err))
			continue
		}

		payload := dtos.StopTaskTickets{TicketIDs: info.ticketIDs}
		url := poolURL + "/api/v1/ticket/cancelPods"

		_, err = httpclient.SendRequest(url, payload, "PATCH")
		if err != nil {
			cancelErrors = append(cancelErrors, fmt.Errorf("pool %s cancel: %w", poolID, err))
			continue
		}

		// Only add to statusUpdates if cancellation was successful
		for _, ticketID := range info.ticketIDs {
			statusUpdates[ticketID] = models.StatusReady
		}
	}

	// If any cancellation failed, return error without updating status or deleting task
	if len(cancelErrors) > 0 {
		log.Printf("failed to cancel all tickets: %v", cancelErrors)
		return apiError.NewInternalServerError("Failed to cancel all tickets, some resources may still be active")
	}

	// All cancellations succeeded, now proceed with cleanup
	if len(statusUpdates) > 0 {
		if err := u.ticketRepository.BatchUpdateTicketStatuses(statusUpdates); err != nil {
			log.Printf("failed to batch update ticket statuses: %v", err)
			return apiError.NewInternalServerError("Failed to update ticket statuses")
		}
	}

	err = u.ticketRepository.DeleteQueue(taskID)
	if err != nil {
		log.Printf("failed to delete queue for task %s: %v", taskID, err)
		return apiError.NewInternalServerError("Failed to clean up task queue")
	}

	// Update task status to cancelled before deletion
	err = u.ticketRepository.UpdateTaskStatus(taskID, models.StatusCancelled)
	if err != nil {
		log.Printf("failed to update task status for task %s: %v", taskID, err)
		return apiError.NewInternalServerError("Failed to update task status")
	}

	err = u.ticketRepository.DeleteTask(taskID)
	if err != nil {
		log.Printf("failed to delete task %s: %v", taskID, err)
		return apiError.NewInternalServerError("Failed to delete task")
	}

	u.requeue(u.getNodeNamesByTickets(task.Tickets))

	return nil
}

func (u *TicketUsecase) StopTask(userID uuid.UUID, taskID uuid.UUID) (interface{}, error) {
	task, err := u.ticketRepository.GetTasksByID(taskID)
	if err != nil {
		log.Printf("failed to get task by ID %s: %v", taskID, err)
		return nil, apiError.NewInternalServerError("Failed to find the task")
	}
	if task.OwnerID != userID {
		return nil, apiError.NewForbiddenError("You do not have permission to stop this task")
	}

	ticketsByPool, err := u.groupTicketByPool(task.Tickets)
	if err != nil {
		return nil, err
	}

	statusUpdates := make(map[uuid.UUID]models.StatusTicket)
	var allResponses []dtos.StopTaskResponse
	for poolID, poolTickets := range ticketsByPool {
		poolURN, err := u.getPoolURN(poolID, poolTickets[0].GlideletURN)
		if err != nil {
			log.Printf("failed to get pool URN for pool %s: %v", poolID, err)
			return nil, apiError.NewInternalServerError("Failed to connect to resource pool")
		}

		url := poolURN + "/api/v1/ticket/deletePods"
		payload := dtos.StopTaskTickets{TicketIDs: []uuid.UUID{}}
		for _, t := range poolTickets {
			payload.TicketIDs = append(payload.TicketIDs, t.GliderTicketID)
		}

		body, err := httpclient.SendRequest(url, payload, "POST")
		if err != nil {
			log.Printf("failed to stop tickets in pool %s: %v", poolID, err)
			return nil, apiError.NewInternalServerError("Failed to stop tickets in resource pool")
		}

		var response []dtos.StopTaskResponse
		if err := json.Unmarshal(body, &response); err != nil {
			log.Printf("failed to parse stop task response from pool %s: %v", poolID, err)
			return nil, apiError.NewInternalServerError("Failed to process stop task response")
		}

		for _, res := range response {
			ticketStatus := models.StatusFailed
			if strings.ToLower(res.Status) == "deleted" {
				ticketStatus = models.StatusStopped
			}
			statusUpdates[res.TicketID] = ticketStatus
		}

		allResponses = append(allResponses, response...)
	}

	if len(statusUpdates) > 0 {
		log.Printf("[STOP TASK] Updating ticket statuses: %v", statusUpdates)
		if err := u.ticketRepository.BatchUpdateTicketStatuses(statusUpdates); err != nil {
			log.Printf("failed to batch update ticket statuses: %v", err)
			return nil, apiError.NewInternalServerError("Failed to update ticket statuses")
		}
	}

	log.Printf("[STOP TASK] Update task status to stopped for task %s", taskID)
	if err := u.updateTaskStatus(taskID); err != nil {
		return nil, err
	}

	u.requeue(u.getNodeNamesByTickets(task.Tickets))

	return allResponses, nil
}

func (u *TicketUsecase) updateTicketInfo(codeServerResponse *dtos.CodeServerResponse) error {
	for _, res := range codeServerResponse.TicketResponse {
		if err := u.ticketRepository.UpdateCodeServerInfo(res.TicketID, res.URL, res.Password); err != nil {
			log.Printf("[UPDATE TICKET INFO] failed to update ticket %s: %v", res.TicketID, err)
			continue
		}
	}

	return nil
}

func (u *TicketUsecase) confirmTickets(ticketsByPool map[uuid.UUID][]dtos.TicketReq) error {
	for poolID, poolTickets := range ticketsByPool {
		poolURN, err := u.getPoolURN(poolID.String(), poolTickets[0].GlideletURN)
		if err != nil {
			log.Printf("failed to get pool URN for pool %s: %v", poolID, err)
			return apiError.NewInternalServerError("Failed to connect to resource pool")
		}

		request := dtos.StopTaskTickets{
			TicketIDs: []uuid.UUID{},
		}
		for _, t := range poolTickets {
			request.TicketIDs = append(request.TicketIDs, t.ID)
		}

		url := poolURN + "/api/v1/ticket/confirmJobs"
		_, err = httpclient.SendRequest(url, request, "PATCH")
		if err != nil {
			log.Printf("failed to confirm tickets with pool %s: %v", poolID, err)
			return apiError.NewInternalServerError("Failed to confirm tickets with resource pool")
		}
	}

	return nil
}

func (u *TicketUsecase) groupTicketIDsByPool(ticketIDs []uuid.UUID) (map[uuid.UUID][]dtos.TicketReq, error) {
	ticketsByPool := make(map[uuid.UUID][]dtos.TicketReq)

	for _, gliderTicketID := range ticketIDs {
		ticketReq, err := u.toTicketRequest(gliderTicketID)
		if err != nil {
			log.Printf("failed to convert ticket to ticket request: %v", err)
			return nil, apiError.NewInternalServerError("Failed to process ticket")
		}

		poolID := ticketReq.Spec.PoolID
		ticketsByPool[poolID] = append(ticketsByPool[poolID], *ticketReq)

	}

	return ticketsByPool, nil
}

func (u *TicketUsecase) groupTicketByPool(tickets []models.Ticket) (map[string][]models.Ticket, error) {
	ticketsByPool := make(map[string][]models.Ticket)

	for _, ticket := range tickets {
		poolID := ticket.ResourcePoolID.String()
		ticketsByPool[poolID] = append(ticketsByPool[poolID], ticket)
	}

	return ticketsByPool, nil
}

func (u *TicketUsecase) getPoolURN(poolID string, defaultURL string) (string, error) {
	url := os.Getenv("CLEARINGHOUSE_URL") + "/resources/pool/" + poolID
	response, err := httpclient.SendRequest(url, nil, "GET")

	if err != nil {
		return defaultURL, nil
	}

	var poolInfo struct {
		GlideletURN string `json:"glidelet_urn"`
	}
	if err := json.Unmarshal(response, &poolInfo); err != nil {
		return defaultURL, nil
	}

	if poolInfo.GlideletURN == "" {
		return defaultURL, nil
	}

	return poolInfo.GlideletURN, nil
}

func (u *TicketUsecase) toTicketRequest(ticketID uuid.UUID) (*dtos.TicketReq, error) {
	ticket, err := u.ticketRepository.GetTicketByGliderTicketID(ticketID)
	if err != nil {
		log.Printf("failed to get ticket by GliderTicketID %s: %v", ticketID, err)
		return nil, apiError.NewInternalServerError("Failed to retrieve ticket")
	}

	ticketReq, err := mapper.ToTicketRequest(ticket)
	if err != nil {
		log.Printf("failed to convert ticket to ticket request: %v", err)
		return nil, apiError.NewInternalServerError("Failed to process ticket")
	}

	return ticketReq, nil
}

func (u *TicketUsecase) DeleteTasks(request dtos.DeleteTasksRequest, userID uuid.UUID) error {
	// validate tickets belong to user
	tickets, err := u.ticketRepository.GetTasksByIDs(request.TaskIDs)
	if err != nil {
		log.Printf("failed to retrieve tasks: %v", err)
		return apiError.NewInternalServerError("Failed to retrieve tasks")
	}
	for _, ticket := range tickets {
		if ticket.OwnerID != userID {
			return apiError.NewForbiddenError("You do not have permission to delete one or more tasks")
		}
	}

	// delete tickets
	if err := u.ticketRepository.DeleteTasksByIDs(request.TaskIDs); err != nil {
		log.Printf("failed to delete tasks: %v", err)
		return apiError.NewInternalServerError("Failed to delete tasks, please try again")
	}
	return nil
}
