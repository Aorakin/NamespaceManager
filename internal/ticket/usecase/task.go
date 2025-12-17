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

	codeServerResponse, err := u.sendTickets(request.Tickets)
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

func (u *TicketUsecase) sendTickets(ticketIDs []uuid.UUID) (*dtos.CodeServerResponse, error) {
	ticketsByPool, err := u.groupTicketsByPool(ticketIDs)
	if err != nil {
		return nil, err
	}

	var allResponses []dtos.TicketResponse
	for poolID, poolTickets := range ticketsByPool {
		poolURN, err := u.getPoolURN(poolID, poolTickets[0].GlideletURN)
		if err != nil {
			return nil, apiError.NewInternalServerError(fmt.Errorf("failed to get pool URN for pool %s: %w", poolID, err))
		}

		url := poolURN + "/api/v1/ticket/createList"
		response, err := httpclient.SendRequest(url, poolTickets, "POST")
		if err != nil {
			return nil, apiError.NewInternalServerError(fmt.Errorf("failed to send tickets to pool %s: %w", poolID, err))
		}

		var poolResponse dtos.CodeServerResponse
		if err := json.Unmarshal(response, &poolResponse); err != nil {
			return nil, apiError.NewBadRequestError(fmt.Errorf("failed to unmarshal response from pool %s: %w", poolID, err))
		}

		allResponses = append(allResponses, poolResponse.TicketResponse...)
	}

	return &dtos.CodeServerResponse{TicketResponse: allResponses}, nil
}

func (u *TicketUsecase) groupTicketsByPool(ticketIDs []uuid.UUID) (map[string][]dtos.TicketReq, error) {
	ticketsByPool := make(map[string][]dtos.TicketReq)

	for _, gliderTicketID := range ticketIDs {
		ticketReq, err := u.toTicketRequest(gliderTicketID)
		if err != nil {
			return nil, err
		}

		poolID := ticketReq.Spec.PoolID.String()
		ticketsByPool[poolID] = append(ticketsByPool[poolID], *ticketReq)
	}

	return ticketsByPool, nil
}

func (u *TicketUsecase) getPoolURN(poolID string, defaultURL string) (string, error) {
	url := os.Getenv("CLEARINGHOUSE_URL") + "/resource/pool/" + poolID
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
