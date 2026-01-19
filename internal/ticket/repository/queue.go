package repository

import (
	"time"

	"github.com/NamespaceManager/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *TicketRepository) GetNextStartTimeByPoolID(poolID uuid.UUID) (time.Time, error) {
	var queueTicket models.QueueTicket
	err := r.db.Where("pool_id = ? AND start_time >= ?", poolID, time.Now()).Order("start_time desc").First(&queueTicket).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}

	return queueTicket.StartTime, nil
}

func (r *TicketRepository) CreateQueueTicket(queueTicket models.QueueTicket) error {
	if err := r.db.Create(&queueTicket).Error; err != nil {
		return err
	}
	return nil
}
