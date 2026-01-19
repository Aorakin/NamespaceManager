package repository

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

func (r *TicketRepository) GetTasks(ownerID uuid.UUID) ([]models.Task, error) {
	var tasks []models.Task
	if err := r.db.Preload("Tickets").Where("owner_id = ?", ownerID).Order("created_at desc").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *TicketRepository) GetTasksByID(taskID uuid.UUID) (*models.Task, error) {
	var task models.Task
	if err := r.db.Preload("Tickets").First(&task, "id = ?", taskID).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TicketRepository) GetNextQueueTask() (*models.Task, error) {
	var task models.Task
	if err := r.db.Preload("Tickets").Where("status = ?", models.StatusQueued).Order("estimated_start_time asc").First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

func (r *TicketRepository) GetQueuedTasks() ([]models.Task, error) {
	var tasks []models.Task
	if err := r.db.Preload("Tickets").Where("status = ?", models.StatusQueued).Order("estimated_start_time asc").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}
