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
		return nil, apiError.NewInternalServerError(fmt.Errorf("failed to get tasks: %w", err))
	}

	return tasks, nil
}

func (u *TicketUsecase) CreateTask(request *dtos.CreateTaskRequest, userID uuid.UUID) error {
	tickets, err := u.ticketRepository.GetTicketsByGliderTicketIDs(request.Tickets)
	if err != nil {
		return apiError.NewInternalServerError(fmt.Errorf("failed to retrieve tickets: %w", err))
	}

	namespaceID := uuid.Nil

	for _, ticket := range tickets {
		if namespaceID == uuid.Nil {
			namespaceID = ticket.NamespaceID
		} else if ticket.NamespaceID != namespaceID {
			return apiError.NewBadRequestError(fmt.Errorf("all tickets must belong to the same namespace"))
		}

		if ticket.Status != models.StatusReady {
			return apiError.NewBadRequestError(fmt.Errorf("ticket %s is not in 'Ready' status", ticket.GliderTicket.ID))
		}
		if ticket.OwnerID != userID {
			return apiError.NewForbiddenError(fmt.Errorf("ticket %s does not belong to the user", ticket.GliderTicket.ID))
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
		return apiError.NewInternalServerError(fmt.Errorf("failed to get task by ID: %w", err))
	}
	if task.OwnerID != userID {
		return apiError.NewForbiddenError(fmt.Errorf("user does not own the task"))
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

	for poolID, info := range ticketsByPool {
		poolURL, err := u.getPoolURN(poolID.String(), info.defaultURL)
		if err != nil {
			return apiError.NewInternalServerError(fmt.Errorf("failed to get pool URN for pool %s: %w", poolID, err))
		}

		payload := dtos.StopTaskTickets{TicketIDs: info.ticketIDs}
		url := poolURL + "/api/v1/ticket/cancelPods"

		_, err = httpclient.SendRequest(url, payload, "PATCH")
		if err != nil {
			return apiError.NewInternalServerError(fmt.Errorf("failed to cancel tickets in pool %s: %w", poolID, err))
		}
		for _, ticketID := range info.ticketIDs {
			statusUpdates[ticketID] = models.StatusReady
		}

	}

	if len(statusUpdates) > 0 {
		if err := u.ticketRepository.BatchUpdateTicketStatuses(statusUpdates); err != nil {
			return apiError.NewInternalServerError(fmt.Errorf("failed to batch update ticket statuses: %w", err))
		}
	}

	err = u.ticketRepository.DeleteQueue(taskID)
	if err != nil {
		return apiError.NewInternalServerError(fmt.Errorf("failed to delete queue: %w", err))
	}

	err = u.ticketRepository.DeleteTask(taskID)
	if err != nil {
		return apiError.NewInternalServerError(fmt.Errorf("failed to delete task: %w", err))
	}

	u.requeue(u.getNodeNamesByTickets(task.Tickets))

	return nil
}

func (u *TicketUsecase) StopTask(userID uuid.UUID, taskID uuid.UUID) (interface{}, error) {
	task, err := u.ticketRepository.GetTasksByID(taskID)
	if err != nil {
		return nil, apiError.NewInternalServerError(fmt.Errorf("failed to get task by ID: %w", err))
	}
	if task.OwnerID != userID {
		return nil, apiError.NewForbiddenError(fmt.Errorf("user does not own the task"))
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
			return nil, apiError.NewInternalServerError(fmt.Errorf("failed to get pool URN for pool %s: %w", poolID, err))
		}

		url := poolURN + "/api/v1/ticket/deletePods"
		payload := dtos.StopTaskTickets{TicketIDs: []uuid.UUID{}}
		for _, t := range poolTickets {
			payload.TicketIDs = append(payload.TicketIDs, t.GliderTicketID)
		}

		body, err := httpclient.SendRequest(url, payload, "POST")
		if err != nil {
			return nil, apiError.NewInternalServerError(fmt.Errorf("failed to stop tickets in pool %s: %w", poolID, err))
		}

		var response []dtos.StopTaskResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, apiError.NewBadRequestError(fmt.Errorf("failed to parse stop task response from pool %s: %w", poolID, err))
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
		if err := u.ticketRepository.BatchUpdateTicketStatuses(statusUpdates); err != nil {
			return nil, apiError.NewInternalServerError(fmt.Errorf("failed to batch update ticket statuses: %w", err))
		}
	}

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
			return apiError.NewInternalServerError(fmt.Errorf("failed to get pool URN for pool %s: %w", poolID, err))
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
			return apiError.NewInternalServerError(fmt.Errorf("failed to get code server info from pool %s: %w", poolID, err))
		}
	}

	return nil
}

func (u *TicketUsecase) groupTicketIDsByPool(ticketIDs []uuid.UUID) (map[uuid.UUID][]dtos.TicketReq, error) {
	ticketsByPool := make(map[uuid.UUID][]dtos.TicketReq)

	for _, gliderTicketID := range ticketIDs {
		ticketReq, err := u.toTicketRequest(gliderTicketID)
		if err != nil {
			return nil, apiError.NewInternalServerError(fmt.Errorf("failed to convert ticket to ticket request: %w", err))
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
		return nil, apiError.NewInternalServerError(fmt.Errorf("failed to get ticket by GliderTicketID: %w", err))
	}

	ticketReq, err := mapper.ToTicketRequest(ticket)
	if err != nil {
		return nil, apiError.NewInternalServerError(fmt.Errorf("failed to convert ticket to ticket request: %w", err))
	}

	return ticketReq, nil
}

func (u *TicketUsecase) DeleteTasks(request dtos.DeleteTasksRequest, userID uuid.UUID) error {
	// validate tickets belong to user
	tickets, err := u.ticketRepository.GetTasksByIDs(request.TaskIDs)
	if err != nil {
		return apiError.NewInternalServerError(fmt.Errorf("failed to get tickets: %w", err))
	}
	for _, ticket := range tickets {
		if ticket.OwnerID != userID {
			return apiError.NewForbiddenError(fmt.Errorf("ticket %s does not belong to user", ticket.ID))
		}
	}

	// delete tickets
	if err := u.ticketRepository.DeleteTasksByIDs(request.TaskIDs); err != nil {
		return apiError.NewInternalServerError(fmt.Errorf("failed to delete tickets: %w", err))
	}
	return nil
}
