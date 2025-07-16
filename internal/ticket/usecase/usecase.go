package usecase

import (
	"encoding/json"
	"fmt"

	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/ticket/dtos"
	"github.com/NamespaceManager/internal/ticket/interfaces"
	"github.com/NamespaceManager/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type TicketUsecase struct {
	TicketRepository interfaces.TicketRepository
}

func NewTicketUsecase(ticketRepository interfaces.TicketRepository) interfaces.TicketUsecase {
	return &TicketUsecase{TicketRepository: ticketRepository}
}

func (u *TicketUsecase) HandleTicketCallback(ticket models.GliderTicket) error {
	validate := validator.New()
	if err := validate.Struct(ticket); err != nil {
		return err
	}

	if err := u.TicketRepository.Create(ticket); err != nil {
		return err
	}
	return nil
}

func (u *TicketUsecase) GetTicketNS(userID uuid.UUID, namespaceID uuid.UUID) ([]dtos.TicketResponse, error) {
	tickets, err := u.TicketRepository.GetTicketNS(userID, namespaceID)
	if err != nil {
		return nil, err
	}
	ticketResponses := u.FormatTicketRes(tickets)
	return ticketResponses, nil
}

func (u *TicketUsecase) UpdateStatus(ticketID uuid.UUID, status models.StatusTicket) error {
	return u.TicketRepository.UpdateStatus(ticketID, status)
}

func (u *TicketUsecase) Delete(ticketID uuid.UUID) error {
	return u.TicketRepository.Delete(ticketID)
}

func (u *TicketUsecase) TicketHis(userID uuid.UUID) ([]dtos.TicketResponse, error) {
	tickets, err := u.TicketRepository.TicketHis(userID)
	if err != nil {
		return nil, err
	}
	ticketResponses := u.FormatTicketRes(tickets)
	return ticketResponses, nil
}

func (u *TicketUsecase) UseTicket(ticketIDs []uuid.UUID, userID uuid.UUID) ([]dtos.Payload, error) {
	tickets := make([]dtos.Payload, len(ticketIDs))
	for i, ticketID := range ticketIDs {
		ticket, err := u.TicketRepository.GetTicketByID(ticketID, userID)
		if err != nil {
			return nil, err
		}
		if ticket.Status != "ready" {
			return nil, fmt.Errorf("ticket not ready")
		}
		if ticket.TaskID != uuid.Nil {
			return nil, fmt.Errorf("duplicate ticket in other task")
		}
		payload, err := u.SetPayload(ticket)
		if err != nil {
			return nil, err
		}
		tickets[i] = *payload
		//ไม่ได้เช็ค ticket owner id match กับ user id
	}
	return tickets, nil
}

func (u *TicketUsecase) CreateTask(ticketIDs []uuid.UUID, ownerID uuid.UUID) error {
	var tickets []models.GliderTicket
	for _, ticketID := range ticketIDs {
		ticket, err := u.TicketRepository.GetTicketByID(ticketID, ownerID)
		fmt.Println(ticket.TaskID != uuid.Nil)
		if err != nil {
			return err
		}

		tickets = append(tickets, ticket)
	}
	task := models.Tasks{
		Owner_ID: ownerID,
		Tickets:  tickets,
	}
	return u.TicketRepository.CreateTask(task)
}

func (u *TicketUsecase) StopTasks(taskID uuid.UUID) error {
	url := "http://host.docker.internal:5000/api/v1/resourceunit"
	fmt.Println("task id", taskID)
	task, err := u.TicketRepository.GetTasksByID(taskID)
	if err != nil {
		return err
	}
	for _, ticket := range task.Tickets {
		payload, err := u.SetPayload(ticket)
		if err != nil {
			return err
		}
		if _, _, err := utils.SendRequest(url, payload, "DELETE"); err != nil {
			fmt.Println(err)
			return err
		}
	}
	for _, ticket := range task.Tickets {
		u.TicketRepository.UpdateStatus(ticket.ID, "inactive")
	}
	if err := u.TicketRepository.RemoveTasks(taskID); err != nil {
		return err
	}
	return nil

}

func (u *TicketUsecase) GetTasks(ownerID uuid.UUID) ([]models.Tasks, error) {
	return u.TicketRepository.GetTasks(ownerID)
}

func (u *TicketUsecase) SetPayload(ticket models.GliderTicket) (*dtos.Payload, error) {
	payload := dtos.Payload{
		GlideletURN:       ticket.GlideletURN,
		ID:                ticket.ID.String(),
		Lease:             ticket.Lease,
		NamespaceURN:      ticket.NamespaceURN,
		RedeemTimeout:     ticket.RedeemTimeout,
		ReferenceTicketID: ticket.ReferenceTicketID,
		Signature:         ticket.ReferenceTicketID,
		Spec:              ticket.Spec,
	}
	return &payload, nil
}

func (u *TicketUsecase) RollbackFailedTickets(listPayload []dtos.Payload, lastIndex int) error {
	url := "http://host.docker.internal:5000/api/v1/resourceunit"
	for i, payload := range listPayload {
		if lastIndex <= i {
			break
		}
		_, _, err := u.TicketRepository.SendRequest(url, payload, "DELETE")
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *TicketUsecase) ApporveTicket(ticketID uuid.UUID, userID uuid.UUID) (models.GliderTicket, error) {
	ticket, err := u.TicketRepository.GetTicketByID(ticketID, userID)
	if err != nil {
		return models.GliderTicket{}, err
	}
	if ticket.Status != "ready" || ticket.OwnerID != userID {
		return models.GliderTicket{}, fmt.Errorf("ticket not ready")
	}
	//check ticket with passport
	return ticket, nil
}

func (u *TicketUsecase) SendTicket(payload interface{}) (int, []map[string]interface{}, error) {
	url := "http://host.docker.internal:5000/api/v1/resourceunit"
	status, body, err := u.TicketRepository.SendRequest(url, payload, "POST")
	if err != nil {
		return 0, nil, err
	}
	var jsonResponse []map[string]interface{}
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return 0, nil, fmt.Errorf("error : Failed to parse response")
	}
	return status, jsonResponse, nil
}

func (u *TicketUsecase) FormatTicketRes(tickets []models.GliderTicket) []dtos.TicketResponse {
	ticketResponses := make([]dtos.TicketResponse, len(tickets))
	for i, ticket := range tickets {
		ticketResponses[i] = dtos.TicketResponse{
			ID:                ticket.ID,
			OwnerID:           ticket.OwnerID,
			Spec:              ticket.Spec,
			ReferenceTicketID: ticket.ReferenceTicketID,
			RedeemTimeout:     ticket.RedeemTimeout,
			Lease:             ticket.Lease,
			Signature:         ticket.Signature,
		}
	}
	return ticketResponses
}
