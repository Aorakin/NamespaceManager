package http

import (
	"encoding/json"
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
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid request body")))
			return
		}
		userid := utils.GetSession(c, "userID").(uuid.UUID)
		if err := h.ticketUsecase.HandleTicketCallback(ticketreq, userid); err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Ticket created"})
	}
}

func (h *TicketHandlers) GetTasks() gin.HandlerFunc {
	return func(c *gin.Context) {
		ownerID := c.MustGet("userID").(uuid.UUID)
		tasks, err := h.ticketUsecase.GetTasks(ownerID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"tasks": tasks})

	}
}

func (h *TicketHandlers) StopTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		if userID == uuid.Nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewUnauthorizedError("Please log in to continue")))
			return
		}

		taskIDParam := c.Param("task_id")
		taskUUID, err := uuid.Parse(taskIDParam)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid task ID")))
			return
		}

		result, err := h.ticketUsecase.StopTask(userID, taskUUID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func (h *TicketHandlers) UseTickets() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)

		var request dtos.CreateTaskRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid request body")))
			return
		}

		err := h.ticketUsecase.CreateTask(&request, userID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Task created"})
	}
}

func (h *TicketHandlers) RequestTicket() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		accessToken := c.MustGet("accessToken").(string)
		username := c.MustGet("firstname").(string) + " " + c.MustGet("lastname").(string)

		var request dtos.RequestTicketDTO
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid input")))
			return
		}

		gliderTicket, err := h.ticketUsecase.RequestTicket(request, accessToken, userID, username)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}

		// Format and return response
		c.JSON(http.StatusCreated, gin.H{"ticket": gliderTicket.Ticket})
	}
}

func (h *TicketHandlers) GetTicketByNamespaceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		namespaceID := c.Param("namespace_id")

		namespaceUUID, err := uuid.Parse(namespaceID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid namespace ID")))
			return
		}

		tickets, err := h.ticketUsecase.GetTicketByNamespaceID(namespaceUUID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to retrieve tickets")))
			return
		}
		c.JSON(http.StatusOK, mapper.ToUserTicketResponseList(tickets))
	}
}

func (h *TicketHandlers) GetTicketByNamespaceIDAndNodeID() gin.HandlerFunc {
	return func(c *gin.Context) {
		namespaceID := c.Param("namespace_id")
		nodeID := c.Param("node_id")

		namespaceUUID, err := uuid.Parse(namespaceID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid namespace ID")))
			return
		}

		nodeUUID, err := uuid.Parse(nodeID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid node ID")))
			return
		}

		tickets, err := h.ticketUsecase.GetTicketByNamespaceIDAndNodeID(namespaceUUID, nodeUUID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to retrieve tickets")))
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
			c.JSON(response.ErrorResponseBuilder(apiError.NewInternalServerError("Failed to retrieve tickets")))
			return
		}

		c.JSON(http.StatusOK, mapper.ToUserTicketResponseList(tickets))
	}
}

func (h *TicketHandlers) UpdateTicketStatusFromGlidelet() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req []dtos.StatusRes
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		if jsonData, err := json.MarshalIndent(req, "", "  "); err == nil {
			log.Println(string(jsonData))
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
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Ticket ID is required")))
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

func (h *TicketHandlers) CancelTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		if userID == uuid.Nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewUnauthorizedError("Please log in to continue")))
			return
		}

		taskIDParam := c.Param("task_id")
		taskUUID, err := uuid.Parse(taskIDParam)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid task ID")))
			return
		}

		err = h.ticketUsecase.CancelTask(userID, taskUUID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}

		c.JSON(http.StatusOK, nil)
	}
}

func (h *TicketHandlers) DeleteTickets() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		if userID == uuid.Nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewUnauthorizedError("Please log in to continue")))
			return
		}

		var req dtos.DeleteTicketsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid request body")))
			return
		}

		err := h.ticketUsecase.DeleteTickets(req, userID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "tickets deleted successfully"})
	}
}

func (h *TicketHandlers) DeleteTasks() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uuid.UUID)
		if userID == uuid.Nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewUnauthorizedError("Please log in to continue")))
			return
		}

		var req dtos.DeleteTasksRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(response.ErrorResponseBuilder(apiError.NewBadRequestError("Invalid request body")))
			return
		}

		err := h.ticketUsecase.DeleteTasks(req, userID)
		if err != nil {
			c.JSON(response.ErrorResponseBuilder(err))
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "tasks deleted successfully"})
	}
}
