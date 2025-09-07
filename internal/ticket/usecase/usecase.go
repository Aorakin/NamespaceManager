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

func (u *TicketUsecase) HandleTicketCallback(ticketreq dtos.CreateTicket, userid uuid.UUID) error { //use for test (ใช้จริงคือสร้างจากที่รับมาจาก CH)
	validate := validator.New()
	if err := validate.Struct(ticketreq); err != nil {
		return err
	}
	ticket := CreateTicketToGliderTicket(ticketreq, userid)
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

func (u *TicketUsecase) UseTicket(ticketIDs []uuid.UUID, userID uuid.UUID) ([]models.GliderTicket, error) {
	tickets := make([]models.GliderTicket, len(ticketIDs))
	for i, ticketID := range ticketIDs {
		ticket, err := u.TicketRepository.GetTicketByID(ticketID, userID)
		if err != nil {
			return nil, err
		}
		if ticket.Status != "ready" {
			return nil, fmt.Errorf("ticket not ready")
		}
		if ticket.TaskID != nil && *ticket.TaskID != uuid.Nil {
			return nil, fmt.Errorf("duplicate ticket in other task")
		}
		tickets[i] = ticket
		//ไม่ได้เช็ค ticket owner id match กับ user id
	}
	return tickets, nil
}

func (u *TicketUsecase) CreateTask(taskReq dtos.CreateTaskRequest, ownerID uuid.UUID) error {
	var tickets []models.GliderTicket
	for _, ticketID := range taskReq.Tickets {
		ticket, err := u.TicketRepository.GetTicketByID(ticketID, ownerID)
		if err != nil {
			return err
		}

		tickets = append(tickets, ticket)
	}
	task := models.Tasks{
		Owner_ID: ownerID,
		Title:    taskReq.Title,
		Description: taskReq.Description,
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
	
	// Try to send DELETE requests to external service, but don't fail if service is unavailable
	for _, ticket := range task.Tickets {
		if _, _, err := utils.SendRequest(url, ticket.ID, "DELETE"); err != nil {
			fmt.Printf("Warning: Failed to send DELETE request to external service: %v\n", err)
			// Continue execution instead of returning error
		}
	}
	
	// First, clear the TaskID from tickets to break the foreign key relationship
	for _, ticket := range task.Tickets {
		if err := u.TicketRepository.ClearTaskID(ticket.ID); err != nil {
			fmt.Printf("Warning: Failed to clear task ID for ticket %s: %v\n", ticket.ID, err)
		}
	}
	
	// Update ticket statuses to inactive
	for _, ticket := range task.Tickets {
		if err := u.TicketRepository.UpdateStatus(ticket.ID, "inactive"); err != nil {
			fmt.Printf("Warning: Failed to update ticket status to inactive for ticket %s: %v\n", ticket.ID, err)
		}
	}
	
	// Finally, remove the task from database
	if err := u.TicketRepository.RemoveTasks(taskID); err != nil {
		return err
	}
	return nil
}

func (u *TicketUsecase) GetTasks(ownerID uuid.UUID) ([]models.Tasks, error) {
	return u.TicketRepository.GetTasks(ownerID)
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
			Status:            string(ticket.Status),
			CreatedAt:        ticket.CreatedAt,
			UpdatedAt:        ticket.UpdatedAt,
		}
	}
	return ticketResponses
}

func CreateTicketToGliderTicket(req dtos.CreateTicket, ownerID uuid.UUID) *models.GliderTicket {
	ticketID := uuid.New()

	gliderSpecs := make([]models.GliderSpec, 0, len(req.Spec))
	for _, specReq := range req.Spec {
		specID := uuid.New()
		specResources := make([]models.SpecResource, 0, len(specReq.Resources))

		for _, resReq := range specReq.Resources {
			specResources = append(specResources, models.SpecResource{
				ID:       uuid.New(),
				Name:     resReq.Name,
				Quantity: resReq.Quantity,
				Unit:     resReq.Unit,
				SpecID:   specID,
			})
		}

		gliderSpecs = append(gliderSpecs, models.GliderSpec{
			ID:        specID,
			TicketID:  ticketID,
			Type:      specReq.Type,
			PoolID:    specReq.PoolID,
			Resources: specResources,
		})
	}

	return &models.GliderTicket{
		ID:                ticketID,
		OwnerID:           ownerID,
		NamespaceID:       req.NamespaceID,
		NamespaceURN:      req.NamespaceURN,
		GlideletURN:       req.GlideletURN,
		Spec:              gliderSpecs,
		ReferenceTicketID: req.ReferenceTicketID,
		RedeemTimeout:     req.RedeemTimeout,
		Lease:             req.Lease,
		Signature:         req.Signature,
	}
}
