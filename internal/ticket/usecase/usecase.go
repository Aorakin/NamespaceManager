package usecase

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/NamespaceManager/internal/models"
	namespaceInterfaces "github.com/NamespaceManager/internal/namespace/interfaces"
	"github.com/NamespaceManager/internal/ticket/dtos"
	"github.com/NamespaceManager/internal/ticket/interfaces"
	userInterfaces "github.com/NamespaceManager/internal/users/interfaces"
	"github.com/NamespaceManager/internal/utils"
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

// func (u *TicketUsecase) GetTicketFromCH(ticketId string) (int, dtos.GliderTicketResponse, error) {
// 	url := fmt.Sprintf("%s/tickets/%s", os.Getenv("CLEARINGHOUSE_URL"), ticketId)

// 	status, body, err := utils.SendRequest(url, nil, "GET")
// 	if err != nil {
// 		return 0, dtos.GliderTicketResponse{}, err
// 	}
// 	var ticket dtos.GliderTicketResponse
// 	err = json.Unmarshal(body, &ticket)
// 	if err != nil {
// 		return 0, dtos.GliderTicketResponse{}, fmt.Errorf("error : Failed to parse response: %w", err)
// 	}

//		return status, ticket, nil
//	}

func (u *TicketUsecase) StopTask(userID uuid.UUID, taskID uuid.UUID) (interface{}, error) {
	task, err := u.ticketRepository.GetTasksByID(taskID)
	if err != nil {
		return nil, err
	}
	if task.OwnerID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	var stopTaskPayload dtos.StopTaskTickets
	for _, ticket := range task.Tickets {
		stopTaskPayload.TicketIDs = append(stopTaskPayload.TicketIDs, ticket.GliderTicket.ID)
	}

	url := os.Getenv("GLIDELET_URL") + ":9443" + "/api/v1/ticket/deletePods"

	status, body, err := utils.SendRequest(url, stopTaskPayload, "POST")
	if err != nil {
		return nil, err
	}

	if status > 299 || status < 200 {
		return body, fmt.Errorf("failed to stop task, status code: %d, response: %s", status, string(body))
	}

	var response []dtos.StopTaskResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse stop task response: %w", err)
	}

	for _, res := range response {
		ticketStatus := models.StatusFailed
		if strings.ToLower(res.Status) == "deleted" {
			ticketStatus = models.StatusStopped
		}
		err := u.ticketRepository.UpdateTicketStatus(res.TicketID, ticketStatus)
		if err != nil {
			return nil, fmt.Errorf("failed to update ticket %s status: %w", res.TicketID, err)
		}
	}

	return body, err
}
