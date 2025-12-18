package http

import (
	"fmt"
	"log"
	"net/http"

	"github.com/NamespaceManager/internal/ticket/dtos"
	"github.com/NamespaceManager/internal/ticket/interfaces"
	"github.com/NamespaceManager/internal/ticket/mapper"
	"github.com/NamespaceManager/internal/utils"
	apiError "github.com/NamespaceManager/pkg/api_error"
	"github.com/NamespaceManager/pkg/response"
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

func (h *TicketHandlers) GetTasks() gin.HandlerFunc {
	return func(c *gin.Context) {
		ownerID := c.MustGet("userID").(uuid.UUID)
		tasks, err := h.ticketUsecase.GetTasks(ownerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{"tasks": tasks})

	}
}

func (h *TicketHandlers) StopTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		if userID == uuid.Nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		taskIDParam := c.Param("task_id")
		taskUUID, err := uuid.Parse(taskIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task_id"})
			return
		}

		response, err := h.ticketUsecase.StopTask(userID, taskUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

func (h *TicketHandlers) UseTickets() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		var request dtos.CreateTaskRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError(err)))
			return
		}

		err := h.ticketUsecase.CreateTask(&request, userID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}
		c.JSON(http.StatusOK, "task created")
	}
}

func (h *TicketHandlers) RequestTicket() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		accessToken := c.MustGet("accessToken").(string)
		username := c.MustGet("firstname").(string) + " " + c.MustGet("lastname").(string)

		var request dtos.RequestTicketDTO
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		gliderTicket, err := h.ticketUsecase.RequestTicket(request, accessToken, userID, username)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}

		// Format and return response
		c.JSON(http.StatusCreated, gin.H{"ticket": gliderTicket})
	}
}

func (h *TicketHandlers) GetTicketByNamespaceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		namespaceID := c.Param("namespace_id")
		namespaceUUID, err := uuid.Parse(namespaceID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError(fmt.Errorf("invalid namespace_id: %w", err))))
			return
		}

		tickets, err := h.ticketUsecase.GetTicketByNamespaceID(namespaceUUID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError(err)))
			return
		}
		c.JSON(http.StatusOK, mapper.ToUserTicketResponseList(tickets))
	}
}

func (h *TicketHandlers) GetUserTickets() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		tickets, err := h.ticketUsecase.GetUserTickets(userID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError(err)))
			return
		}

		c.JSON(http.StatusOK, mapper.ToUserTicketResponseList(tickets))
	}
}

func (h *TicketHandlers) UpdateTicketStatusFromGlidelet() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req []dtos.StatusRes
		log.Println(req)
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		err := h.ticketUsecase.UpdateTicketStatusFromGlidelet(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Ticket statuses updated successfully"})
	}
}

func (h *TicketHandlers) CancelTicket() gin.HandlerFunc {
	return func(c *gin.Context) {
		ticketId := c.Param("ticket_id")
		accessToken := c.MustGet("accessToken").(string)

		if ticketId == "" {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError(fmt.Errorf("ticket_id is required"))))
			return
		}

		err := h.ticketUsecase.CancelTicket(ticketId, accessToken)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "ticket cancelled successfully"})
	}
}
