package usecase

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/ticket/dtos"
	"github.com/NamespaceManager/internal/ticket/interfaces"
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
	// ticket := CreateTicketToGliderTicket(ticketreq, userid)
	// if err := u.TicketRepository.Create(ticket); err != nil {
	// 	return err
	// }
	return nil
}

func (u *TicketUsecase) UseTicket(ticketIDs []uuid.UUID) ([]models.GliderTicket, error) {
	tickets := make([]models.GliderTicket, len(ticketIDs))
	// for i, ticketID := range ticketIDs {
	// 	ticket, err := u.TicketRepository.GetTicketByID(ticketID)
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

func (u *TicketUsecase) CreateTask(taskReq dtos.CreateTaskRequest, ownerID uuid.UUID) error {
	var tickets []models.Ticket
	var ticketIDs []uuid.UUID

	for _, ticketID := range taskReq.Tickets {
		ticket, err := u.TicketRepository.GetTicketByGliderTicketID(ticketID)
		if err != nil {
			return err
		}
		tickets = append(tickets, ticket)
		ticketIDs = append(ticketIDs, ticketID)
	}

	// Send tickets to external service
	status, response, err := u.SendTicket(ticketIDs)
	if err != nil {
		return fmt.Errorf("failed to send tickets: %w", err)
	}
	log.Printf("Ticket send response: %v", response)
	// You can check the status and response if needed
	if status != 200 {
		return fmt.Errorf("failed to send tickets, status code: %d", status)
	}

	task := models.Task{
		Title:   taskReq.Title,
		Tickets: tickets,
		OwnerID: ownerID,
		Status:  models.StatusPending,
	}

	return u.TicketRepository.CreateTask(task)
}

func (u *TicketUsecase) StopTask(taskID uuid.UUID) error {
	// url := "http://host.docker.internal:5000/api/v1/resourceunit"
	// fmt.Println("task id", taskID)
	task, err := u.TicketRepository.GetTasksByID(taskID)
	if err != nil {
		return err
	}

	// // Try to send DELETE requests to external service, but don't fail if service is unavailable
	// for _, ticket := range task.Tickets {
	// 	if _, _, err := utils.SendRequest(url, ticket.ID, "DELETE"); err != nil {
	// 		fmt.Printf("Warning: Failed to send DELETE request to external service: %v\n", err)
	// 		// Continue execution instead of returning error
	// 	}
	// }

	// // First, clear the TaskID from tickets to break the foreign key relationship
	for _, ticket := range task.Tickets {
		if err := u.TicketRepository.ClearTaskID(ticket.ID); err != nil {
			fmt.Printf("Warning: Failed to clear task ID for ticket %s: %v\n", ticket.ID, err)
		}
	}

	// // Update ticket statuses to inactive
	for _, ticket := range task.Tickets {
		if err := u.TicketRepository.UpdateTicketStatus(ticket.GliderTicket.ID, models.StatusStopped); err != nil {
			fmt.Printf("Warning: Failed to update ticket status to stopped for ticket %s: %v\n", ticket.ID, err)
		}
	}

	// // Finally, remove the task from database
	if err := u.TicketRepository.UpdateTaskStatus(taskID, models.StatusStopped); err != nil {
		return err
	}
	return nil
}

func (u *TicketUsecase) GetTasks(ownerID uuid.UUID) ([]models.Task, error) {
	return u.TicketRepository.GetTasks(ownerID)
}

// func (u *TicketUsecase) RollbackFailedTickets(listPayload []dtos.Payload, lastIndex int) error {
// 	url := "http://host.docker.internal:5000/api/v1/resourceunit"
// 	for i, payload := range listPayload {
// 		if lastIndex <= i {
// 			break
// 		}
// 		_, _, err := u.TicketRepository.SendRequest(url, payload, "DELETE")
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func (u *TicketUsecase) ConvertTicketToTicketRequest(t models.Ticket) (dtos.TicketReq, error) {
	// Convert poolID (string) to uuid.UUID
	poolID, err := uuid.Parse(t.GliderTicket.Spec.PoolID)
	if err != nil {
		return dtos.TicketReq{}, fmt.Errorf("invalid pool_id: %w", err)
	}

	// Convert resources
	var resources []dtos.SpecResource
	for _, r := range t.GliderTicket.Spec.Resources {
		resources = append(resources, dtos.SpecResource{
			Name:     r.Name,
			Quantity: int64(r.Quantity),
			Unit:     r.Unit,
		})
	}

	req := dtos.TicketReq{
		GlideletURN:       t.GliderTicket.GlideletURN,
		ID:                t.GliderTicket.ID,
		Lease:             fmt.Sprintf("%d", t.GliderTicket.Lease),
		NamespaceURN:      t.GliderTicket.NamespaceURN,
		RedeemTimeout:     fmt.Sprintf("%d", t.GliderTicket.RedeemTimeout),
		ReferenceTicketID: t.GliderTicket.ReferenceTicketID,
		Signature:         t.Signature,
		Spec: dtos.GliderSpec{
			Type:      dtos.ResourceUnitType(t.GliderTicket.Spec.Type),
			PoolID:    poolID,
			Resources: resources,
		},
	}

	return req, nil
}

func (u *TicketUsecase) ApproveTicket(ticketID uuid.UUID) (dtos.TicketReq, error) {
	ticket, err := u.TicketRepository.GetTicketByGliderTicketID(ticketID)
	if err != nil {
		return dtos.TicketReq{}, err
	}
	ticketReq, err := u.ConvertTicketToTicketRequest(ticket)
	if err != nil {
		return dtos.TicketReq{}, err
	}
	return ticketReq, nil
}

func (u *TicketUsecase) SendTicket(payload []uuid.UUID) (int, *dtos.CodeServerResponse, error) {
	var tickets []dtos.TicketReq
	fmt.Println("payload", payload)
	for _, ticketID := range payload {
		ticket, err := u.ApproveTicket(ticketID)
		if err != nil {
			return 0, nil, err
		}
		tickets = append(tickets, ticket)
	}
	fmt.Println(tickets)
	url := os.Getenv("GLIDELET_URL") + ":9443" + "/api/v1/ticket/createList"
	status, body, err := u.TicketRepository.SendRequest(url, tickets, "POST")
	log.Printf("status: %d, body: %s", status, string(body))

	if err != nil {
		return 0, nil, err
	}
	var jsonResponse dtos.CodeServerResponse
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return 0, nil, fmt.Errorf("error : Failed to parse response: %w", err)
	}
	log.Printf("response: %v", jsonResponse)

	for _, res := range jsonResponse.TicketResponse {
		ticket, err := u.TicketRepository.GetTicketByGliderTicketID(res.TicketID)
		if err != nil {
			return 0, nil, err
		}
		ticket.URL = res.URL
		ticket.Password = res.Password
		err = u.TicketRepository.Update(ticket)
		if err != nil {
			return 0, nil, err
		}
	}

	return status, &jsonResponse, nil
}

// func (u *TicketUsecase) SetPayload(ticket models.GliderTicket) (*dtos.Payload, error) {
// 	payload := dtos.Payload{
// 		GlideletURN:       ticket.GlideletURN,
// 		ID:                ticket.ID.String(),
// 		Lease:             ticket.Lease,
// 		NamespaceURN:      ticket.NamespaceURN,
// 		RedeemTimeout:     ticket.RedeemTimeout,
// 		ReferenceTicketID: ticket.ReferenceTicketID,
// 		Signature:         ticket.ReferenceTicketID,
// 		Spec:              ticket.Spec,
// 	}
// 	return &payload, nil
// }

// func (u *TicketUsecase) FormatTicketRes(tickets []models.GliderTicket) []dtos.TicketResponse {
// 	ticketResponses := make([]dtos.TicketResponse, len(tickets))
// 	// for i, ticket := range tickets {
// 	// 	ticketResponses[i] = dtos.TicketResponse{
// 	// 		ID:                ticket.ID,
// 	// 		OwnerID:           ticket.OwnerID,
// 	// 		Spec:              ticket.Spec,
// 	// 		ReferenceTicketID: ticket.ReferenceTicketID,
// 	// 		RedeemTimeout:     ticket.RedeemTimeout,
// 	// 		Lease:             ticket.Lease,
// 	// 		Signature:         ticket.Signature,
// 	// 		Status:            string(ticket.Status),
// 	// 		CreatedAt:         ticket.CreatedAt,
// 	// 		UpdatedAt:         ticket.UpdatedAt,
// 	// 	}
// 	// }
// 	return ticketResponses
// }

func (u *TicketUsecase) TicketModeltoDTO(tickets []models.Ticket) []dtos.UserTicketResponse {
	var dtosList []dtos.UserTicketResponse

	for _, t := range tickets {
		dto := dtos.UserTicketResponse{
			ID:           t.ID.String(),
			Name:         t.Name,
			GliderTicket: models.GliderTicket(t.GliderTicket),
			Signature:    t.Signature,
			Status:       string(t.Status),
		}
		dtosList = append(dtosList, dto)
	}

	return dtosList
}

func (u *TicketUsecase) GetTicketByNamespaceID(namespaceId string) ([]dtos.UserTicketResponse, error) {
	tickets, err := u.TicketRepository.GetTicketByNamespaceID(namespaceId)
	if err != nil {
		return []dtos.UserTicketResponse{}, err
	}
	if len(tickets) == 0 {
		return []dtos.UserTicketResponse{}, nil
	}
	ticketResponses := u.TicketModeltoDTO(tickets)
	return ticketResponses, nil
}

func (u *TicketUsecase) GetUserTickets(ownerID uuid.UUID) ([]dtos.UserTicketResponse, error) {
	tickets, err := u.TicketRepository.GetUserTickets(ownerID)
	if err != nil {
		return []dtos.UserTicketResponse{}, err
	}
	if len(tickets) == 0 {
		return []dtos.UserTicketResponse{}, nil
	}
	ticketResponses := u.TicketModeltoDTO(tickets)
	return ticketResponses, nil
}

//	func (u *TicketUsecase) RequestTicketToCH(ticketReq dtos.RequestTicketDTO) (int, dtos.GliderTicketResponse, error) {
//		url := os.Getenv("CLEARINGHOUSE_URL") + "/tickets/"
//		status, res, err := utils.SendRequest(url, ticketReq, "POST")
//		fmt.Println("error", err)
//		fmt.Println("status", status)
//		fmt.Println("res", string(res))
//		if err != nil {
//			return 0, dtos.GliderTicketResponse{}, err
//		}
//		var resTicket dtos.GliderTicketResponse
//		if err := json.Unmarshal(res, &resTicket); err != nil {
//			return 0, dtos.GliderTicketResponse{}, err
//		}
//		return status, resTicket, nil
//	}

func (u *TicketUsecase) SaveTicket(ticketRes dtos.GliderTicketResponse, name string, ownerID uuid.UUID) error {
	ticket := models.Ticket{
		Name:         name,
		GliderTicket: models.GliderTicketJSON(ticketRes.Ticket),
		Signature:    ticketRes.Signature,
		Status:       models.StatusReady,
		OwnerID:      ownerID,
		TaskID:       nil,
	}

	if err := u.TicketRepository.Create(&ticket); err != nil {
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

// 	return status, ticket, nil
// }

func (u *TicketUsecase) UpdateTaskStatus(userID uuid.UUID, taskID uuid.UUID) error {
	task, err := u.TicketRepository.GetTasksByID(taskID)
	if err != nil {
		return err
	}
	if task.OwnerID != userID {
		return fmt.Errorf("unauthorized")
	}

	tickets, err := u.TicketRepository.GetTicketsByTaskID(taskID)
	if err != nil {
		return err
	}

	var taskStatus models.StatusTicket
	allRedeemed := true
	anyPending := false

	for _, ticket := range tickets {
		if ticket.Status == models.StatusPending {
			anyPending = true
			allRedeemed = false
			break
		}
		if ticket.Status != models.StatusRedeemed {
			allRedeemed = false
		}
	}

	if allRedeemed {
		taskStatus = models.StatusRedeemed
	} else if anyPending {
		taskStatus = models.StatusPending
	} else {
		taskStatus = models.StatusFailed
	}

	return u.TicketRepository.UpdateTaskStatus(taskID, taskStatus)

}

func (u *TicketUsecase) UpdateTicketStatusFromGlidelet(req []dtos.StatusRes) error {
	for _, statusRes := range req {
		if statusRes.HasError {
			err := u.TicketRepository.UpdateTicketStatus(statusRes.TicketID, models.StatusFailed)
			if err != nil {
				fmt.Printf("failed to update ticket %s to status failed: %v", statusRes.TicketID, err)
			}
			continue
		}

		if len(statusRes.PodStatus) == 0 {
			fmt.Printf("Warning: No pod status found for ticketId: %s. Skipping update.", statusRes.TicketID)
			continue
		}

		var finalTicketStatus models.StatusTicket
		ticketPending := false
		ticketRunning := true

		for _, pod := range statusRes.PodStatus {
			if pod.Status == "pending" {
				ticketPending = true
				break
			}

			if pod.Status != "running" {
				ticketRunning = false
			}
		}

		if ticketPending {
			finalTicketStatus = models.StatusPending
		} else if ticketRunning {
			finalTicketStatus = models.StatusRedeemed
		} else {
			fmt.Printf("Info: TicketId %s has an indeterminate status (not all running, none pending). Skipping update.", statusRes.TicketID)
			continue
		}

		fmt.Printf("Updating ticket %s to status %s", statusRes.TicketID, finalTicketStatus)
		err := u.TicketRepository.UpdateTicketStatus(statusRes.TicketID, finalTicketStatus)
		if err != nil {
			fmt.Printf("failed to update ticket %s to status %s: %v", statusRes.TicketID, finalTicketStatus, err)
		}
	}

	return nil
}

func (u *TicketUsecase) CancelTicket(ticketID string) error {
	// url := os.Getenv("CLEARINGHOUSE_URL") + "/tickets/" + ticketID + "/cancel"
	// status, _, err := utils.SendRequest(url, nil, "PATCH")
	// if err != nil {
	// 	return err
	// }
	// if status != 200 {
	// 	return fmt.Errorf("failed to cancel ticket in Clearinghouse, status code: %d", status)
	// }
	err := u.TicketRepository.CancelTicket(ticketID)
	if err != nil {
		return err
	}
	return nil
}
