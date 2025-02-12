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

func (r *TicketRepository) Create(ticket models.GliderTicket) error {
	if err := r.db.Create(&ticket).Error; err != nil {
		return err
	}
	return nil
}

func (r *TicketRepository) GetMyTicket(userID uuid.UUID, namespaceID uuid.UUID) ([]models.GliderTicket, error) {
	var tickets []models.GliderTicket
	if err := r.db.Preload("Spec.Resources").
		Where("owner_id = ? AND namespace_id = ?", userID, namespaceID).
		Find(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

func (r *TicketRepository) GetTicketByID(ID uuid.UUID) (*models.GliderTicket, error) {
	var ticket models.GliderTicket
	if err := r.db.Preload("Spec.Resources").First(&ticket, "id = ?", ID).Error; err != nil {
		return nil, err
	}

	if len(ticket.Spec) == 0 {
		return nil, fmt.Errorf("ticket must have at least one spec")
	}

	return &ticket, nil
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
