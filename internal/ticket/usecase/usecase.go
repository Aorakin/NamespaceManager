package usecase

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/NamespaceManager/internal/models"
	namespaceInterfaces "github.com/NamespaceManager/internal/namespace/interfaces"
	"github.com/NamespaceManager/internal/ticket/dtos"
	"github.com/NamespaceManager/internal/ticket/interfaces"
	userInterfaces "github.com/NamespaceManager/internal/users/interfaces"
	apiError "github.com/NamespaceManager/pkg/api_error"
	"github.com/NamespaceManager/pkg/httpclient"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type TicketUsecase struct {
	ticketRepository interfaces.TicketRepository
	namespaceRepo    namespaceInterfaces.NamespaceRepository
	userRepo         userInterfaces.UsersRepository
}

func NewTicketUsecase(ticketRepository interfaces.TicketRepository, namespaceRepo namespaceInterfaces.NamespaceRepository, userRepo userInterfaces.UsersRepository) interfaces.TicketUsecase {
	return &TicketUsecase{ticketRepository: ticketRepository, namespaceRepo: namespaceRepo, userRepo: userRepo}
}

func (u *TicketUsecase) HandleTicketCallback(ticketreq dtos.CreateTicket, userid uuid.UUID) error { //use for test (ใช้จริงคือสร้างจากที่รับมาจาก CH)
	validate := validator.New()
	if err := validate.Struct(ticketreq); err != nil {
		return err
	}
	// ticket := CreateTicketToGliderTicket(ticketreq, userid)
	// if err := u.ticketRepository.Create(ticket); err != nil {
	// 	return err
	// }
	return nil
}

func (u *TicketUsecase) UseTicket(ticketIDs []uuid.UUID) ([]models.GliderTicket, error) {
	tickets := make([]models.GliderTicket, len(ticketIDs))
	// for i, ticketID := range ticketIDs {
	// 	ticket, err := u.ticketRepository.GetTicketByID(ticketID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if ticket.Status != "ready" {
	// 		return nil, fmt.Errorf("ticket not ready")
	// 	}
	// 	if ticket.TaskID != nil && *ticket.TaskID != uuid.Nil {
	// 		return nil, fmt.Errorf("duplicate ticket in other task")
	// 	}
	// 	tickets[i] = ticket
	// 	//ไม่ได้เช็ค ticket owner id match กับ user id
	// }
	return tickets, nil
}

func (u *TicketUsecase) SaveTicket(ticketRes dtos.GliderTicketResponse, name string, ownerID uuid.UUID) error {
	ticket := models.Ticket{
		Name:         name,
		GliderTicket: ticketRes.Ticket,
		Signature:    ticketRes.Signature,
		Status:       models.StatusReady,
		OwnerID:      ownerID,
		TaskID:       nil,
	}

	if err := u.ticketRepository.Create(&ticket); err != nil {
		return err
	}

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
	var allResponses []dtos.StopTaskResponse

	for poolID, info := range ticketsByPool {
		poolURL, err := u.getPoolURN(poolID.String(), info.defaultURL)
		if err != nil {
			return nil, apiError.NewInternalServerError(fmt.Errorf("failed to get pool URN for pool %s: %w", poolID, err))
		}

		payload := dtos.StopTaskTickets{TicketIDs: info.ticketIDs}
		url := poolURL + "/api/v1/ticket/deletePods"

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

	return allResponses, nil
}
