package repository

import (
	"github.com/NamespaceManager/internal/models"
	"github.com/google/uuid"
)

func (r *TicketRepository) GetUserTickets(userID uuid.UUID) ([]models.Ticket, error) {
	var tickets []models.Ticket
	if err := r.db.Where("owner_id = ?", userID).Order("created_at DESC").Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

func (r *TicketRepository) GetTicketsByGliderTicketIDs(ticketIDs []uuid.UUID) ([]models.Ticket, error) {
	var tickets []models.Ticket
	if err := r.db.Where("glider_ticket->> 'id' IN ?", ticketIDs).Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

func (r *TicketRepository) UpdateCodeServerInfo(ticketID uuid.UUID, url string, password string) error {
	var ticket models.Ticket
	if err := r.db.Where("glider_ticket->> 'id' = ?", ticketID).First(&ticket).Error; err != nil {
		return err
	}

	ticket.URL = url
	ticket.Password = password

	if err := r.db.Save(&ticket).Error; err != nil {
		return err
	}

	return nil
}
