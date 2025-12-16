package usecase

import (
	"encoding/json"
	"fmt"
	"os"

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

	for _, task := range tasks {
		if err := u.updateTaskStatus(task.ID); err != nil {
			return nil, apiError.NewInternalServerError(fmt.Errorf("failed to update task status for task %s: %v", task.ID, err))
		}
	}

	tasks, err = u.ticketRepository.GetTasks(ownerID)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (u *TicketUsecase) CreateTask(request *dtos.CreateTaskRequest, userID uuid.UUID) error {
	tickets, err := u.ticketRepository.GetTicketsByGliderTicketIDs(request.Tickets)
	if err != nil {
		return apiError.NewInternalServerError(fmt.Errorf("failed to retrieve tickets: %w", err))
	}

	for _, ticket := range tickets {
		if ticket.Status != models.StatusReady {
			return apiError.NewBadRequestError(fmt.Errorf("ticket %s is not in 'Ready' status", ticket.GliderTicket.ID))
		}
		if ticket.OwnerID != userID {
			return apiError.NewForbiddenError(fmt.Errorf("ticket %s does not belong to the user", ticket.GliderTicket.ID))
		}
	}

	codeServerResponse, err := u.sendTicket(request.Tickets)
	if err != nil {
		return err
	}

	if err := u.updateTicketInfo(codeServerResponse); err != nil {
		return err
	}

	task := models.Task{
		Title:   request.Title,
		Tickets: tickets,
		OwnerID: userID,
		Status:  models.StatusPending,
	}

	return u.ticketRepository.CreateTask(task)
}

func (u *TicketUsecase) updateTicketInfo(codeServerResponse *dtos.CodeServerResponse) error {
	for _, res := range codeServerResponse.TicketResponse {
		if err := u.ticketRepository.UpdateCodeServerInfo(res.TicketID, res.URL, res.Password); err != nil {
			return fmt.Errorf("failed to update ticket info: %w", err)
		}
	}

	return nil
}

func (u *TicketUsecase) sendTicket(tickets []uuid.UUID) (*dtos.CodeServerResponse, error) {
	var ticketsReq []dtos.TicketReq
	for _, gliderTicketID := range tickets {
		ticket, err := u.toTicketRequest(gliderTicketID)
		if err != nil {
			return nil, err
		}
		ticketsReq = append(ticketsReq, *ticket)
	}

	url := os.Getenv("GLIDELET_URL") + ":9443" + "/api/v1/ticket/createList"
	response, err := httpclient.SendRequest(url, ticketsReq, "POST")
	if err != nil {
		return nil, err
	}

	var codeServerResponse dtos.CodeServerResponse
	err = json.Unmarshal(response, &codeServerResponse)
	if err != nil {
		return nil, apiError.NewBadRequestError(fmt.Errorf("failed to unmarshal response: %w", err))
	}

	return &codeServerResponse, nil
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

func (u *TicketUsecase) updateTaskStatus(taskID uuid.UUID) error {
	tickets, err := u.ticketRepository.GetTicketsByTaskID(taskID)
	if err != nil {
		return err
	}

	var (
		anyFailed   bool
		anyStopped  bool
		anyPending  bool
		anyExpired  bool
		allRedeemed = true
	)

	for _, t := range tickets {
		switch t.Status {
		case models.StatusFailed:
			anyFailed = true
			allRedeemed = false

		case models.StatusStopped:
			anyStopped = true
			allRedeemed = false

		case models.StatusPending:
			anyPending = true
			allRedeemed = false

		case models.StatusExpired:
			anyExpired = true
			allRedeemed = false

		case models.StatusRedeemed:
			// still possibly all redeemed
		default:
			allRedeemed = false
		}
	}

	var taskStatus models.StatusTicket

	// Apply your priority:
	switch {
	case anyExpired:
		taskStatus = models.StatusExpired
	case anyFailed:
		taskStatus = models.StatusFailed
	case anyStopped:
		taskStatus = models.StatusStopped
	case anyPending:
		taskStatus = models.StatusPending
	case allRedeemed:
		taskStatus = models.StatusRedeemed
	default:
		taskStatus = models.StatusFailed
	}

	return u.ticketRepository.UpdateTaskStatus(taskID, taskStatus)
}
