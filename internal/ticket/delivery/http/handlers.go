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
		var ticketreq dtos.CreateTicket
		if err := c.ShouldBindJSON(&ticketreq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userid := utils.GetSession(c, "userID").(uuid.UUID)
		if err := h.ticketUsecase.HandleTicketCallback(ticketreq, userid); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, "Ticket Created")
	}
}

func (h *TicketHandlers) GetHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := utils.GetSession(c, "userID").(uuid.UUID)
		tickets, err := h.ticketUsecase.TicketHis(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"tickets": tickets})
	}
}

func (h *TicketHandlers) SendTicket() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ticketReq dtos.TicketIDRequest
		if err := c.ShouldBindJSON(&ticketReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		userID := utils.GetSession(c, "userID").(uuid.UUID)
		ticket, err := h.ticketUsecase.ApporveTicket(ticketReq.ID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ticket not match"})
			return
		}

		payload, err := h.ticketUsecase.SetPayload(ticket)
		if err != nil || ticket.Status != "ready" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		status, jsonResponse, err := h.ticketUsecase.SendTicket(payload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		if err := h.ticketUsecase.UpdateStatus(ticket.ID, "active"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		c.JSON(status, jsonResponse)
	}
}

func (h *TicketHandlers) GetTicketNS() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ticketReq dtos.RequestWithNS
		if err := c.ShouldBindJSON(&ticketReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		userID := utils.GetSession(c, "userID").(uuid.UUID)

		tickets, err := h.ticketUsecase.GetTicketNS(userID, ticketReq.NamespaceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{"tickets": tickets})
	}
}

func (h *TicketHandlers) Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dtos.UpdateTicketStatusReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		if err := h.ticketUsecase.UpdateStatus(req.ID, req.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{"Ticket": "updated"})
	}
}

func (h *TicketHandlers) Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req uuid.UUID
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		if err := h.ticketUsecase.Delete(req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{"Ticket": "deleted"})
	}
}

func (h *TicketHandlers) GetTasks() gin.HandlerFunc {
	return func(c *gin.Context) {
		ownerID := utils.GetSession(c, "userID").(uuid.UUID)
		tasks, err := h.ticketUsecase.GetTasks(ownerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{"Tasks": tasks})

	}
}

func (h *TicketHandlers) RemoveTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dtos.RequestTaskID
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		if err := h.ticketUsecase.StopTasks(req.TaskID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Task stopped successfully"})
	}
}

func (h *TicketHandlers) UseTickets() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ticketReq []uuid.UUID
		if err := c.ShouldBindJSON(&ticketReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		userID := utils.GetSession(c, "userID").(uuid.UUID)

		// listPayload, err := h.ticketUsecase.UseTicket(ticketReq, userID)
		// if err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// 	return
		// }
		// resp := make([][]map[string]interface{}, len(listPayload))
		// var successfulPayloads []dtos.Payload
		// var laststatus int
		// for i, payload := range listPayload {
		// 	status, jsonResponse, err := h.ticketUsecase.SendTicket(payload)
		// 	if err != nil {
		// 		if err := h.ticketUsecase.RollbackFailedTickets(successfulPayloads, i); err != nil {
		// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		// 		}
		// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// 		return
		// 	}
		// 	successfulPayloads = append(successfulPayloads, payload)
		// 	resp[i] = jsonResponse
		// 	laststatus = status
		// }
		if err := h.ticketUsecase.CreateTask(ticketReq, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
	 	c.JSON(http.StatusOK, "response sent to CH and task created successfully")
	}
}

func (h *TicketHandlers) RequestTicket() gin.HandlerFunc { // request ticket to CH but never test
	return func(c *gin.Context) {
		var ticketReq dtos.RequestTicket
		if err := c.ShouldBindJSON(&ticketReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		if utils.CheckValidater(ticketReq) != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		// userID := utils.GetSession(c, "userID").(uuid.UUID)

		url := "http://host.docker.internal:8989" //change url
		status, jsonResponse, err := utils.SendRequest(url, ticketReq, "POST")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var tickets []models.GliderTicket
		if err := json.Unmarshal(jsonResponse, &tickets); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
			return
		}
		// for _, ticket := range tickets {
		// 	if err := h.ticketUsecase.HandleTicketCallback(ticket, userID); err != nil {
		// 		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		// 		return
		// 	}
		// }
		c.JSON(status, tickets)

	}
}
