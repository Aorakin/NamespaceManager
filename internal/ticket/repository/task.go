package repository

import (
	"time"

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

func (r *TicketRepository) DeleteTask(taskID uuid.UUID) error {
	// First, set TaskID to NULL for all tickets associated with this task
	if err := r.db.Model(&models.Ticket{}).Where("task_id = ?", taskID).Update("task_id", nil).Error; err != nil {
		return err
	}

	// Then delete the task only if it has a deletable status
	result := r.db.Where("id = ? AND status IN ?", taskID, models.UneditableStatus).Delete(&models.Task{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *TicketRepository) GetTasksByNodeNames(nodeNames []string) ([]models.Task, error) {
	var tasks []models.Task
	err := r.db.
		Distinct("tasks.*").
		Model(&models.Task{}).
		Preload("Tickets").
		Joins("JOIN tickets ON tickets.task_id = tasks.id").
		Where("tasks.status = ?", models.StatusQueued).
		Where("tickets.glider_ticket->>'node_name' IN ?", nodeNames).
		Order("tasks.created_at asc").
		Find(&tasks).Error

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *TicketRepository) UpdateTaskQueueInfo(taskID uuid.UUID, startTime time.Time, status models.StatusTicket) error {
	if err := r.db.Model(&models.Task{}).Where("id = ?", taskID).Updates(models.Task{EstimatedStartTime: startTime, Status: status}).Error; err != nil {
		return err
	}
	return nil
}

func (r *TicketRepository) GetTasksByIDs(taskIDs []uuid.UUID) ([]models.Task, error) {
	var tasks []models.Task
	if err := r.db.Where("id IN ?", taskIDs).Order("created_at asc").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *TicketRepository) DeleteTasksByIDs(taskIDs []uuid.UUID) error {
	result := r.db.Where("id IN ? AND status IN ?", taskIDs, models.UneditableStatus).Delete(&models.Task{})
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *TicketRepository) UpdateTaskStartTime(taskID uuid.UUID, startTime time.Time) error {
	if err := r.db.Model(&models.Task{}).Where("id = ?", taskID).Update("started_at", startTime).Error; err != nil {
		return err
	}
	return nil
}
