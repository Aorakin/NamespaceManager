package repository

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/google/uuid"
)

func (r *TicketRepository) GetUserTickets(userID uuid.UUID) ([]models.Ticket, error) {
	var tickets []models.Ticket
	if err := r.db.Where("owner_id = ?", userID).Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}
