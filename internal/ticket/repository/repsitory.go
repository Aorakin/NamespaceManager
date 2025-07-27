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
	var ticket models.GliderTicket
	if err := r.db.First(&ticket, ticketID).Error; err != nil {
		return err
	}

	ticket.Status = updatedData

	if err := r.db.Save(&ticket).Error; err != nil {
		return err
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

func (r *TicketRepository) GetTicketByID(ID uuid.UUID, userID uuid.UUID) (models.GliderTicket, error) {
	var ticket models.GliderTicket
	if err := r.db.Preload("Spec.Resources").
		Where("id = ? AND owner_id = ?", ID, userID).
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
