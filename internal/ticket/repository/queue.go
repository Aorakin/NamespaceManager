package repository

import (
	"time"

	"github.com/NamespaceManager/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *TicketRepository) GetNextStartTime(poolID uuid.UUID, NodeNames []string) (time.Time, error) {
	var queueTicket models.QueueTicket
	err := r.db.Where("pool_id = ? AND start_time >= ? AND node_name IN ?", poolID, time.Now(), NodeNames).Order("start_time desc").First(&queueTicket).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}

	return queueTicket.StartTime, nil
}

func (r *TicketRepository) GetHeadTask() (*models.Task, error) {
	var task models.Task
	if err := r.db.Preload("Tickets").Where("status = ? AND estimated_start_time IS NOT NULL", models.StatusQueued).Order("estimated_start_time asc").First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

func (r *TicketRepository) CreateQueueTicket(queueTicket models.QueueTicket) error {
	if err := r.db.Create(&queueTicket).Error; err != nil {
		return err
	}
	return nil
}

// GetHeadTasksByPoolAndNodes returns the earliest queued ticket for each node in the pool
func (r *TicketRepository) GetHeadTasksByPoolAndNodes(poolID uuid.UUID, nodeNames []string) (map[string]*models.QueueTicket, error) {
	result := make(map[string]*models.QueueTicket)

	for _, nodeName := range nodeNames {
		var queueTicket models.QueueTicket
		err := r.db.Where("pool_id = ? AND node_name = ? AND start_time >= ?", poolID, nodeName, time.Now()).
			Order("start_time asc").
			First(&queueTicket).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// This node has no head task
				result[nodeName] = nil
				continue
			}
			return nil, err
		}
		result[nodeName] = &queueTicket
	}

	return result, nil
}

func (r *TicketRepository) DeleteQueue(taskID uuid.UUID) error {
	if err := r.db.Where("glider_ticket_id IN (?)",
		r.db.Table("tickets").Select("glider_ticket_id").Where("task_id = ?", taskID),
	).Delete(&models.QueueTicket{}).Error; err != nil {
		return err
	}
	return nil
}
