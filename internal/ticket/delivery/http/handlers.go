package http

import (
	"encoding/json"
	"net/http"

	"github.com/NamespaceManager/internal/models"
	"github.com/NamespaceManager/internal/ticket/dtos"
	"github.com/NamespaceManager/internal/ticket/interfaces"
	"github.com/NamespaceManager/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TicketHandlers struct {
	ticketUsecase interfaces.TicketUsecase
}

func NewTicketHandler(ticketUsecase interfaces.TicketUsecase) interfaces.TicketHandler {
	return &TicketHandlers{ticketUsecase: ticketUsecase}
}

func (h *TicketHandlers) HandleTicketCallback() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ticket models.GliderTicket
		if err := c.ShouldBindJSON(&ticket); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		if err := h.ticketUsecase.HandleTicketCallback(ticket); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, "Ticket Created")
	}
}

func (h *TicketHandlers) SendTicket() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ticketReq dtos.TicketIDRequest
		if err := c.ShouldBindJSON(&ticketReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		ticket, err := h.ticketUsecase.ApporveTicket(ticketReq.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ticket not match"})
			return
		}

		payload, err := h.ticketUsecase.SetPayload(ticket)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		url := "http://host.docker.internal:5000/api/v1/resourceunit"
		status, jsonResponse, err := h.ticketUsecase.SendTicket(url, *payload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		c.JSON(status, jsonResponse)
	}
}

func (h *TicketHandlers) GetMyTicket() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ticketReq dtos.RequestWithNS
		if err := c.ShouldBindJSON(&ticketReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		userID := utils.GetSession(c, "userID").(uuid.UUID)

		tickets, err := h.ticketUsecase.GetMyTicket(userID, ticketReq.NamespaceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{"tickets": tickets})
	}
}

func (h *TicketHandlers) RequestTicket() gin.HandlerFunc { // request ticket to CH but never test
	return func(c *gin.Context) {
		var ticketReq dtos.RequestTicket
		if err := c.ShouldBindJSON(&ticketReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		url := "http://host.docker.internal:8989" //change url
		status, jsonResponse, err := utils.SendRequest(url, ticketReq, "POST")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		var tickets []models.GliderTicket
		if err := json.Unmarshal(jsonResponse, &tickets); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
			return
		}
		for _, ticket := range tickets {
			if err := h.ticketUsecase.HandleTicketCallback(ticket); err != nil {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
		}
		c.JSON(status, tickets)

	}
}
