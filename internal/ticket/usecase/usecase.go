package usecase

import (
	"encoding/json"
	"fmt"
	"log"
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

func (u *TicketUsecase) CreateTask(taskReq dtos.CreateTaskRequest, ownerID uuid.UUID) error {
	var tickets []models.Ticket
	var ticketIDs []uuid.UUID

	for _, ticketID := range taskReq.Tickets {
		ticket, err := u.ticketRepository.GetTicketByGliderTicketID(ticketID)
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

	return u.ticketRepository.CreateTask(task)
}

func (u *TicketUsecase) GetTasks(ownerID uuid.UUID) ([]models.Task, error) {
	tasks, err := u.ticketRepository.GetTasks(ownerID)
	if err != nil {
		return nil, err
	}
	for _, task := range tasks {
		if err := u.updateTaskStatus(ownerID, task.ID); err != nil {
			return nil, fmt.Errorf("failed to update task status for task %s: %v", task.ID, err)
		}
	}
	tasks, err = u.ticketRepository.GetTasks(ownerID)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// func (u *TicketUsecase) RollbackFailedTickets(listPayload []dtos.Payload, lastIndex int) error {
// 	url := "http://host.docker.internal:5000/api/v1/resourceunit"
// 	for i, payload := range listPayload {
// 		if lastIndex <= i {
// 			break
// 		}
// 		_, _, err := u.ticketRepository.SendRequest(url, payload, "DELETE")
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
	ticket, err := u.ticketRepository.GetTicketByGliderTicketID(ticketID)
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
	status, body, err := u.ticketRepository.SendRequest(url, tickets, "POST")
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
		ticket, err := u.ticketRepository.GetTicketByGliderTicketID(res.TicketID)
		if err != nil {
			return 0, nil, err
		}
		ticket.GlideletURN = res.URL
		ticket.Password = res.Password
		err = u.ticketRepository.Update(ticket)
		if err != nil {
			return 0, nil, err
		}
	}

	return status, &jsonResponse, nil
}

func (u *TicketUsecase) SaveTicket(ticketRes dtos.GliderTicketResponse, name string, ownerID uuid.UUID) error {
	ticket := models.Ticket{
		Name:         name,
		GliderTicket: models.GliderTicketJSON(ticketRes.Ticket),
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
func (u *TicketUsecase) updateTaskStatus(userID uuid.UUID, taskID uuid.UUID) error {
	task, err := u.ticketRepository.GetTasksByID(taskID)
	if err != nil {
		return err
	}
	if task.OwnerID != userID {
		return fmt.Errorf("unauthorized")
	}

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

func (u *TicketUsecase) UpdateTicketStatusFromGlidelet(req []dtos.StatusRes) error {
	b, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		log.Println("json marshal error:", err)
		return err
	}
	log.Println(string(b))
	for _, statusRes := range req {
		if statusRes.HasError {
			err := u.ticketRepository.UpdateTicketStatus(statusRes.TicketID, models.StatusFailed)
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
		ticketRunning := true

		for _, pod := range statusRes.PodStatus {
			if strings.ToLower(pod.Status) == "inactive" {
				finalTicketStatus = models.StatusExpired
				ticketRunning = false
				break
			}

			if strings.ToLower(pod.Status) == "pending" {
				finalTicketStatus = models.StatusPending
				ticketRunning = false
				break
			}

			if strings.ToLower(pod.Status) != "running" {
				ticketRunning = false
			}
		}

		if ticketRunning {
			finalTicketStatus = models.StatusRedeemed
		}

		fmt.Printf("Updating ticket %s to status %s", statusRes.TicketID, finalTicketStatus)
		err := u.ticketRepository.UpdateTicketStatus(statusRes.TicketID, finalTicketStatus)
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
	err := u.ticketRepository.CancelTicket(ticketID)
	if err != nil {
		return err
	}
	return nil
}

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
