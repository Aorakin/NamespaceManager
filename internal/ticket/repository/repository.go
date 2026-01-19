package repository

import (
	"fmt"
	"log"
	"slices"

	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/ticket/interfaces"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketRepository struct {
	db *gorm.DB
}

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
	if err := r.db.Where("glider_ticket_id = ?", ticketID).First(&ticket).Error; err != nil {
		return err
	}

	ticket.Status = models.StatusCancelled
	if err := r.db.Save(&ticket).Error; err != nil {
		return err
	}

	return nil
}

func (r *TicketRepository) GetTicketByNamespaceID(namespaceID uuid.UUID) ([]models.Ticket, error) {
	var tickets []models.Ticket
	if err := r.db.Where("namespace_id = ?", namespaceID).Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

func (r *TicketRepository) Update(ticket models.Ticket) error {
	if err := r.db.Save(&ticket).Error; err != nil {
		return err
	}
	return nil
}

func (r *TicketRepository) UpdateTicketStatus(ticketID uuid.UUID, updatedData models.StatusTicket) error {
	var ticket models.Ticket
	if err := r.db.First(&ticket, "glider_ticket_id = ?", ticketID).Error; err != nil {
		return err
	}

	if slices.Contains(models.UneditableStatus, ticket.Status) {
		return fmt.Errorf("cannot update ticket with status: %s", ticket.Status)
	}

	result := r.db.Model(&models.Ticket{}).Where("glider_ticket_id = ?", ticketID).Update("status", updatedData)
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
	if err := r.db.First(&ticket, "glider_ticket_id = ?", ID).Error; err != nil {
		return models.Ticket{}, err
	}
	return ticket, nil
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

func (r *TicketRepository) GetTicketsByTaskID(taskID uuid.UUID) ([]models.Ticket, error) {
	var tickets []models.Ticket
	if err := r.db.Where("task_id = ?", taskID).Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

// BatchUpdateTicketStatuses updates multiple ticket statuses in a single transaction
func (r *TicketRepository) BatchUpdateTicketStatuses(updates map[uuid.UUID]models.StatusTicket) error {
	if len(updates) == 0 {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		for ticketID, status := range updates {
			// Check if the ticket's current status is expired
			var ticket models.Ticket
			if err := tx.Where("glider_ticket_id = ?", ticketID).First(&ticket).Error; err != nil {
				log.Printf("[BATCH UPDATE TICKET STATUS] failed to fetch ticket %s: %v", ticketID, err)
				continue
			}

			// Skip update if current status is expired
			if ticket.Status == models.StatusExpired {
				continue
			}

			result := tx.Model(&models.Ticket{}).
				Where("glider_ticket_id = ?", ticketID).
				Update("status", status)

			if result.Error != nil {
				log.Printf("[BATCH UPDATE TICKET STATUS] failed to update ticket %s: %v", ticketID, result.Error)
				continue
			}
		}
		return nil
	})
}

func (r *TicketRepository) DeleteTask(taskID uuid.UUID) error {
	// First, set TaskID to NULL for all tickets associated with this task
	if err := r.db.Model(&models.Ticket{}).Where("task_id = ?", taskID).Update("task_id", nil).Error; err != nil {
		return err
	}

	// Then delete the task
	result := r.db.Delete(&models.Task{}, "id = ?", taskID)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
