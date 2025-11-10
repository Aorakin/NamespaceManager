package repository

import (
	"fmt"
	"time"

	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/ticket/interfaces"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketRepository struct {
	db *gorm.DB
}

var client = &http.Client{Timeout: 10 * time.Second}

func NewTicketRepository(db *gorm.DB) interfaces.TicketRepository {
	return &TicketRepository{db: db}
}

func (r *TicketRepository) Create(ticket *models.Ticket) error {
	if err := r.db.Create(&ticket).Error; err != nil {
		return err
	}
	return nil
}

func (r *TicketRepository) CancelTicket(ticketID string) error {
	var ticket models.Ticket
	if err := r.db.Where("glider_ticket->> 'id' = ?", ticketID).First(&ticket).Error; err != nil {
		return err
	}

	ticket.Status = models.StatusCancelled
	if err := r.db.Save(&ticket).Error; err != nil {
		return err
	}

	return nil
}

func (r *TicketRepository) GetTicketByNamespaceID(namespaceID string) ([]models.Ticket, error) {
	var tickets []models.Ticket
	if err := r.db.Where("glider_ticket->>'namespace_urn' = ?", namespaceID).Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

func (r *TicketRepository) GetUserTickets(userID uuid.UUID) ([]models.Ticket, error) {
	var tickets []models.Ticket
	if err := r.db.Where("owner_id = ?", userID).Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

func (r *TicketRepository) CreateTask(task models.Task) error {
	if err := r.db.Create(&task).Error; err != nil {
		return err
	}

	for _, ticket := range task.Tickets {
		if err := r.UpdateTicketStatus(ticket.GliderTicket.ID, models.StatusPending); err != nil {
			return err
		}
	}
	return nil
}
func (r *TicketRepository) UpdateTicketStatus(ticketID uuid.UUID, updatedData models.StatusTicket) error {
	result := r.db.Model(&models.Ticket{}).Where("glider_ticket->> 'id' = ?", ticketID).Update("status", updatedData)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no ticket found with ID: %s", ticketID)
	}
	return nil
}

func (r *TicketRepository) GetTicketByGliderTicketID(ID uuid.UUID) (models.Ticket, error) {
	var ticket models.Ticket
	if err := r.db.First(&ticket, "glider_ticket->> 'id' = ?", ID).Error; err != nil {
		return models.Ticket{}, err
	}
	return ticket, nil
}

func (r *TicketRepository) GetTasksByID(taskID uuid.UUID) (*models.Task, error) {
	var task models.Task
	if err := r.db.Preload("Tickets").First(&task, "id = ?", taskID).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TicketRepository) GetTasks(ownerID uuid.UUID) ([]models.Task, error) {
	var tasks []models.Task
	if err := r.db.Preload("Tickets").Where("owner_id = ?", ownerID).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *TicketRepository) UpdateTaskStatus(taskID uuid.UUID, status models.StatusTicket) error {
	result := r.db.Model(&models.Task{}).Where("id = ?", taskID).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *TicketRepository) ClearTaskID(ticketID uuid.UUID) error {
	// Clear the task_id field to break the foreign key relationship
	result := r.db.Model(&models.GliderTicket{}).Where("id = ?", ticketID).Update("task_id", nil)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *TicketRepository) SendRequest(url string, payload interface{}, method string) (int, []byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("Authorization", "Bearer your token")

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("failed to read response: %w", err)
	}

	return resp.StatusCode, body, nil
}
