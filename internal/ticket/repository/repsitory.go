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

func (r *TicketRepository) Create(ticket *models.GliderTicket) error {
	if err := r.db.Create(&ticket).Error; err != nil {
		return err
	}
	return nil
}

func (r *TicketRepository) GetTicketByNamespaceID(namespaceID string) ([]models.Ticket, error) {
	   var tickets []models.Ticket
    if err := r.db.Where("namespace_id = ?", namespaceID).Find(&tickets).Error; err != nil {
        return nil, err
    }
    if len(tickets) == 0 {
        return tickets, nil
    }
    for i := range tickets {
        if err := r.db.Where("ticket_id = ?", tickets[i].ID).Find(&tickets[i].Resources).Error; err != nil {
            return nil, err
        }
    }
    return tickets, nil
}

func (r *TicketRepository) UpsertTicketFromCH(ticketreq *models.Ticket) error {
    // First, try to find existing ticket by ExternalID
    var existingTicket models.Ticket
    if err := r.db.Preload("Resources").Where("external_id = ?", ticketreq.ExternalID).First(&existingTicket).Error; err == nil {
        // Ticket exists, update it
        ticketreq.ID = existingTicket.ID // Use existing ID
        
        // Delete existing resources first to avoid duplicates
        if err := r.db.Where("ticket_id = ?", existingTicket.ID).Delete(&models.Resource{}).Error; err != nil {
            return err
        }
        
        // Now save the ticket with new resources
        return r.db.Save(ticketreq).Error
    }
    // Ticket doesn't exist, create new one
    return r.db.Create(ticketreq).Error
}

func (r *TicketRepository) GetTicketNS(userID uuid.UUID, namespaceID uuid.UUID) ([]models.GliderTicket, error) {
	var tickets []models.GliderTicket
	if err := r.db.Preload("Spec.Resources").
		Where("owner_id = ? AND namespace_id = ?", userID, namespaceID).
		Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

func (r *TicketRepository) UpdateStatus(ticketID uuid.UUID, updatedData models.StatusTicket) error {
	// Use direct SQL update to avoid foreign key constraint issues
	result := r.db.Model(&models.GliderTicket{}).Where("id = ?", ticketID).Update("status", updatedData)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no ticket found with ID: %s", ticketID)
	}
	return nil
}

func (r *TicketRepository) Delete(ticketID uuid.UUID) error {
	var ticket models.GliderTicket
	if err := r.db.First(&ticket, ticketID).Error; err != nil {
		return err
	}

	if err := r.db.Where("ticket_id = ?", ticketID).Delete(&models.GliderSpec{}).Error; err != nil {
		return err
	}

	if err := r.db.Delete(&ticket).Error; err != nil {
		return err
	}

	return nil
}

func (r *TicketRepository) TicketHis(userID uuid.UUID) ([]models.GliderTicket, error) {
	var tickets []models.GliderTicket
	if err := r.db.Preload("Spec.Resources").Find(&tickets, "owner_id = ?", userID).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

func (r *TicketRepository) GetTicketByID(ID uuid.UUID) (models.GliderTicket, error) {
	var ticket models.GliderTicket
	if err := r.db.Preload("Spec.Resources").
		Where("id = ?", ID).
		First(&ticket, "id = ?", ID).Error; err != nil {
		return models.GliderTicket{}, err
	}

	if len(ticket.Spec) == 0 {
		return models.GliderTicket{}, fmt.Errorf("ticket must have at least one spec")
	}

	return ticket, nil
}

func (r *TicketRepository) CreateTask(task models.Tasks) error {
	if err := r.db.Create(&task).Error; err != nil {
		return err
	}

	for _, ticket := range task.Tickets {
		if err := r.UpdateStatus(ticket.ID, "active"); err != nil {
			return err
		}
	}
	return nil
}

func (r *TicketRepository) GetTasksByID(taskID uuid.UUID) (*models.Tasks, error) {
	var task models.Tasks
	if err := r.db.Preload("Tickets.Spec.Resources").First(&task, "id = ?", taskID).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TicketRepository) GetTasks(ownerID uuid.UUID) ([]models.Tasks, error) {
	var tasks []models.Tasks
	if err := r.db.Preload("Tickets.Spec.Resources").Find(&tasks, "owner_id = ?", ownerID).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *TicketRepository) ClearTaskID(ticketID uuid.UUID) error {
	// Clear the task_id field to break the foreign key relationship
	result := r.db.Model(&models.GliderTicket{}).Where("id = ?", ticketID).Update("task_id", nil)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *TicketRepository) RemoveTasks(taskID uuid.UUID) error {
	if err := r.db.Where("id = ?", taskID).Delete(&models.Tasks{}).Error; err != nil {
		return err
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
